package goRedis

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestOperator(t *testing.T) {
	pool := GetRedisPool()
	pool.Create("crm:friend:sync", &Config{
		Address:   "10.0.0.135:6379",
		Db:        0,
		Password:  "admin",
		MaxActive: 10,
	})

	pool.Create("crm:group:sync", &Config{
		Address:   "10.0.0.135:6379",
		Db:        1,
		Password:  "admin",
		MaxActive: 10,
	})

	friendSync, exist := pool.GetConn("crm:friend:sync")
	if !exist {
		panic("crm:friend:sync not found")
	}

	// 写入一个key
	friendSync.Set(context.Background(), "hello", "world", time.Minute*10)

	time.Sleep(time.Second * 5)
	// 取一下
	value, err := friendSync.Get(context.Background(), "hello").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("value: ", value)
}
