package goRedis

import (
	"context"
	"errors"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redis_rate/v9"
	"sync"
	"time"
)

type RedisPool struct {
	Pool map[string]redis.UniversalClient
}

var IotaRedisPools RedisOperator

func NewIotaRedisPools()  {
	IotaRedisPools = &RedisPool{
		Pool: map[string]redis.UniversalClient{},
	}
}

var iotaRedisPoolsLock sync.Mutex

// RedisOperator redis 多实例操作接口
type RedisOperator interface {
	// Create 初始化一个连接 失败时将panic
	Create(key string, opt *redis.UniversalOptions) error
	// Delete 移除一个连接
	Delete(key string)
	// GetConn 获取一个Redis连接
	GetConn(key string) (redis.UniversalClient, bool)
	// GetCache 使用 github.com/go-redis/cache/v8 库，基于 cache.NewTinyLFU 的本地缓存对象
	// 依据压测: TinyLFU 适合少数key（约100k） TinyLFU的内存开销可以通过第二个参数(过期时间)进行调整。
	// 其他缓存策略 参考: https://github.com/vmihailenco/go-cache-benchmark
	// 使用方法 GetCache().Set(&cache.Item{}) 如果没有初始化则会，初始化Redis本地全局缓存，
	// 默认缓存过期时间 1 day，如果有特殊需要可以不初始化全局缓存，直接新建一个缓存对象进行操作
	GetCache(key string) (*cache.Cache, bool)
	// GetLimiter 基于github.com/go-redis/redis_rate实现的令牌桶限制器
	// 类似于rate.Limiter，使用方法 GetLimiter.Allow(ctx, "project:123", redis_rate.PerSecond(10))
	GetLimiter(key string) (*redis_rate.Limiter, bool)
}

// Create 初始化一个连接 失败时将panic
func (rp *RedisPool) Create(key string, opt *redis.UniversalOptions) error {
	iotaRedisPoolsLock.Lock()

	if _, ok := rp.Pool[key]; !ok {
		client := redis.NewUniversalClient(opt)
		err := client.Ping(context.Background()).Err()
		if err != nil {
			return err
		}
		rp.Pool[key] = client
		rp.Pool[key].AddHook(NewJaegerHook())
	} else {
		return errors.New("redis connect key exist")
	}

	iotaRedisPoolsLock.Unlock()

	return nil
}

// Delete 移除一个连接
func (rp *RedisPool) Delete(key string) {
	iotaRedisPoolsLock.Lock()
	delete(rp.Pool, key)
	iotaRedisPoolsLock.Unlock()
}

// GetConn 获取一个Redis连接
func (rp *RedisPool) GetConn(key string) (redis.UniversalClient, bool) {
	client, exist := rp.Pool[key]
	if !exist {
		return nil, false
	}
	return client, true
}

// GetCache 使用 github.com/go-redis/cache/v8 库，基于 cache.NewTinyLFU 的本地缓存对象
func (rp *RedisPool) GetCache(key string) (*cache.Cache, bool) {
	client, exist := rp.Pool[key]
	if !exist {
		return nil, false
	}
	c := cache.New(&cache.Options{
		Redis:        client.(redis.UniversalClient),
		LocalCache:   cache.NewTinyLFU(1<<10, time.Hour*24),
		StatsEnabled: false,
	})

	return c, true
}

// GetLimiter 基于github.com/go-redis/redis_rate实现的令牌桶限制器
func (rp *RedisPool) GetLimiter(key string) (*redis_rate.Limiter, bool) {
	client, exist := rp.Pool[key]
	if !exist {
		return nil, false
	}
	limiter := redis_rate.NewLimiter(client.(redis.UniversalClient))
	return limiter, true
}
