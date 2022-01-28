package goRedis

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redis_rate/v9"
	"github.com/opentracing/opentracing-go"
	flog "github.com/opentracing/opentracing-go/log"
	"github.com/actorbuf/iota/trace"
)

var (
	_globalRedisPool    redis.UniversalClient
	_globalRedisLimiter *redis_rate.Limiter
	_localRedisCache    *cache.Cache
)

// GetConn 获取全局Redis连接（池），不需要释放
func GetConn() redis.UniversalClient {
	// XXX 这里可能出现类型不为空，值为空的情况，遇到坑请注意检查
	// 通常使用SetRedisPool方法来避免发生这种情况
	if _globalRedisPool == nil {
		panic("redis connections pool is not ready")
	}
	return _globalRedisPool
}

// AddHook 添加钩子
func AddHook(hook redis.Hook) {
	_globalRedisPool.AddHook(hook)
}

// AddTraceHook 添加链路追踪钩子
func AddTraceHook() {
	_globalRedisPool.AddHook(&OpenTracingHook{})
}

// SetRedisPool 设置全局Redis连接池
func SetRedisPool(client redis.UniversalClient) {
	if p := reflect.ValueOf(client); p.Type() == nil || p.IsZero() {
		panic("Set redis pool failed.Client is nil")
	}
	_globalRedisPool = client
}

// GetCache 使用 github.com/go-redis/cache/v8 库，基于 cache.NewTinyLFU 的本地缓存对象
// 依据压测: TinyLFU 适合少数key（约100k） TinyLFU的内存开销可以通过第二个参数(过期时间)进行调整。
// 其他缓存策略 参考: https://github.com/vmihailenco/go-cache-benchmark
// 使用方法 GetCache().Set(&cache.Item{}) 如果没有初始化则会，初始化Redis本地全局缓存，
// 默认缓存过期时间 1 day，如果有特殊需要可以不初始化全局缓存，直接新建一个缓存对象进行操作
func GetCache() *cache.Cache {
	// 懒加载
	if _localRedisCache == nil {
		var client redis.UniversalClient
		if client = GetConn(); client == nil {
			panic("redis is not ready")
		}
		_localRedisCache = cache.New(&cache.Options{
			Redis:        client,
			LocalCache:   cache.NewTinyLFU(1<<9, time.Hour*24),
			StatsEnabled: false,
		})
	}
	return _localRedisCache
}

// GetLimiter 基于github.com/go-redis/redis_rate实现的令牌桶限制器
// 类似于rate.Limiter，使用方法 GetLimiter.Allow(ctx, "project:123", redis_rate.PerSecond(10))
func GetLimiter() *redis_rate.Limiter {
	// 懒加载
	if _globalRedisLimiter == nil {
		var client redis.UniversalClient
		if client = GetConn(); client == nil {
			panic("redis is not ready")
		}
		_globalRedisLimiter = redis_rate.NewLimiter(client)
	}
	return _globalRedisLimiter
}

// OpenTracingHook 链路追踪钩子 - 如果要自定义Redis连接可以直接注入
type OpenTracingHook struct {
	// Deprecated: goroutine not safe
	Span opentracing.Span
	// Deprecated: goroutine not safe
	startAt time.Time
}

const openTracingHookStartKey = "hgr-sa"
const openTracingHookSpanName = "go-redis"

var _ redis.Hook = (*OpenTracingHook)(nil)

func (h *OpenTracingHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	// add start at
	ctx = context.WithValue(ctx, openTracingHookStartKey, time.Now())
	return ctx, nil
}

func (h *OpenTracingHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	// new span to do
	span := trace.ObtainChildSpan(ctx, openTracingHookSpanName)
	defer span.Finish()

	var duration time.Duration
	start, ok := ctx.Value(openTracingHookStartKey).(time.Time)
	if ok {
		duration = time.Since(start)
	}
	span.LogFields(
		flog.Object("args", cmd.Args()),
		flog.String("command", cmd.String()),
		flog.Object("duration", duration),
	)
	return nil
}

func (h *OpenTracingHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	// add start at
	ctx = context.WithValue(ctx, openTracingHookStartKey, time.Now())
	return ctx, nil
}

func (h *OpenTracingHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	// new span to do
	span := trace.ObtainChildSpan(ctx, openTracingHookSpanName)
	defer span.Finish()

	var duration time.Duration
	start, ok := ctx.Value(openTracingHookStartKey).(time.Time)
	if ok {
		duration = time.Since(start)
	}
	for k, cmd := range cmds {
		span.LogFields(
			flog.Object(fmt.Sprintf("%d.args", k), cmd.Args()),
			flog.String(fmt.Sprintf("%d.command", k), cmd.String()),
			flog.Object(fmt.Sprintf("%d.duration", k), duration),
		)
	}
	return nil
}

// 全局对象的实现逻辑

// NewClient 新建一个Redis全局连接池，请注意传入 被初始化一次过后 再次初始化将不会发生任何事情
func NewClient(opt *redis.UniversalOptions) {
	// 已经存在了 不需要再次初始化
	if _globalRedisPool != nil {
		return
	}
	client, err := NewClientForSingle(opt)
	if err != nil {
		panic(err)
	}
	if client == nil {
		panic("connect to redis service failed")
	}
	SetRedisPool(client)
}

// NewClientForSingle 新建一个Redis连接池 请保持该连接 否则用完就释放了
func NewClientForSingle(opt *redis.UniversalOptions) (redis.UniversalClient, error) {
	client := redis.NewUniversalClient(opt)
	if client == nil {
		return nil, errors.New("client is nil")
	}
	err := client.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}
	return client, nil
}
