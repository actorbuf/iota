package goRedis

import (
	"context"
	"fmt"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redis_rate/v9"
	"github.com/actorbuf/iota/trace"
	"github.com/actorbuf/iota/trace/jaeger"
	"math/rand"
	"testing"
	"time"
)

var jaegerConfig = &jaeger.Config{
	ServiceName: "heyWoods",
	AgentHost:   "172.20.58.127:6831",
	LogSpans:    true,
	Disabled:    false,
	SamplerCfg: jaeger.SamplerCfg{
		Type:  jaeger.TypeConst,
		Param: jaeger.DefaultConstParam,
	},
}

func TestGoRedis(t *testing.T) {
	closer, err := jaeger.NewJaeger(jaegerConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer closer.Close()

	NewClient(&redis.UniversalOptions{
		Addrs:    []string{"172.20.58.127:6379"},
		DB:       0,
		Password: "",
		PoolSize: 10,
	})

	for i := 0; i < 10; i++ {
		span1 := trace.New("test1")
		ctx1 := trace.NewTracerContext(context.Background(), span1)
		res, err := GetConn().Set(ctx1, "test1", "ttt", 1*time.Second).Result()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(res)
		span1.Finish()
	}

	span2 := trace.New("test2")
	ctx2 := trace.NewTracerContext(context.Background(), span2)
	_, err = GetConn().Pipelined(ctx2, func(pipe redis.Pipeliner) error {
		for i := 1; i <= 100; i++ {
			pipe.Set(ctx2, fmt.Sprintf("key%d", i), rand.Intn(i), 5*time.Second)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	span2.Finish()
}

func TestGetCache(t *testing.T) {
	closer, err := jaeger.NewJaeger(jaegerConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer closer.Close()

	NewClient(&redis.UniversalOptions{
		Addrs:    []string{"172.20.58.127:6379"},
		DB:       0,
		Password: "",
		PoolSize: 10,
	})

	span := trace.New("set cache")
	ctx := trace.NewTracerContext(context.Background(), span)
	err = GetCache().Set(&cache.Item{
		Ctx:            ctx,
		Key:            "cacheTest1",
		Value:          "11111",
		TTL:            5 * time.Second,
		Do:             nil,
		SetXX:          false,
		SetNX:          false,
		SkipLocalCache: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	span.Finish()

	span = trace.New("get cache")
	ctx = trace.NewTracerContext(context.Background(), span)
	var cacheTest string
	err = GetCache().Get(ctx, "cacheTest1", &cacheTest)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cacheTest)
	cacheTest = ""
	err = GetCache().GetSkippingLocalCache(ctx, "cacheTest1", &cacheTest)
	t.Log(cacheTest)
	span.Finish()

	if err = GetConn().Set(context.Background(), "cacheTest2", "22222", 5*time.Second).Err(); err != nil {
		t.Fatal(err)
	}

	span = trace.New("get cache2")
	ctx = trace.NewTracerContext(context.Background(), span)
	var cacheTest2 string
	for i := 0; i < 100; i++ {
		if err = GetCache().Get(ctx, "cacheTest2", &cacheTest2); err != nil {
			t.Fatal(err)
		}
	}
	span.Finish()
}

func TestGetLimiter(t *testing.T) {
	closer, err := jaeger.NewJaeger(jaegerConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer closer.Close()

	NewClient(&redis.UniversalOptions{
		Addrs:    []string{"172.20.58.127:6379"},
		DB:       0,
		Password: "",
		PoolSize: 10,
	})

	span := trace.New("limiter")
	ctx := trace.NewTracerContext(context.Background(), span)

	// 每分钟10次
	limit := redis_rate.PerMinute(10)
	for i := 0; i < 11; i++ {
		res, err := GetLimiter().Allow(ctx, "limiter1", limit)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("allowed", res.Allowed, "remaining", res.Remaining)
	}

	for {
		res, err := GetLimiter().AllowN(ctx, "limiter1", limit, 3)
		if err != nil {
			t.Fatal(err)
		}
		if res.Allowed == 0 {
			select {
			case <-time.After(res.RetryAfter):
			case <-time.After(3 * time.Minute):
			}
		} else {
			t.Log("allowed", res.Allowed, "remaining", res.Remaining)
			break
		}
	}

	span.Finish()
}
