package goRedis

import (
	"context"
	"fmt"
	"github.com/actorbuf/iota/library/jaeger"
	"github.com/go-redis/redis/v8"
	"testing"
	"time"
)

func TestOperator(t *testing.T) {
	_, _, err := jaeger.NewJaegerTracer(&jaeger.Config{
		"127.0.0.1",
		"6831",
		"const",
		1,
	}, "test")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(jaeger.Tracer)

	NewIotaRedisPools()

	// 不带链路追踪的初始化
	err = IotaRedisPools.Create("redis0", &redis.UniversalOptions{
		Addrs: []string{"127.0.0.1:6380"},
		DB: 0,
	}, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 带链路追踪的初始化
	err = IotaRedisPools.Create("redis1", &redis.UniversalOptions{
		Addrs: []string{"127.0.0.1:6380"},
		DB: 1,
	}, NewJaegerHook())
	if err != nil {
		fmt.Println(err)
		return
	}

	redis1, bool := IotaRedisPools.GetConn("redis1")
	if !bool {
		fmt.Println("获取连接失败", redis1)
		return
	}

	ctx := context.Background()

	// 写入一个key
	redis1.Set(ctx, "hello", "world", time.Minute*10)

	time.Sleep(time.Second * 5)
	// 取一下
	value, err := redis1.Get(ctx, "hello").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("value: ", value)

	time.Sleep(time.Second*10)
}
