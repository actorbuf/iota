package common

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"time"

	"github.com/go-redis/redis/v8"
)

type redisLock struct {
	redis   redis.UniversalClient // redis实例
	lockKey string                // 锁key
	token   string                // 锁的值
	ttl     int                   // 秒级ttl
	wait    bool                  // 是否等待
	step    int                   // 尝试锁次数
}

// NewLock 初始化锁
func NewLock(redis redis.UniversalClient) *redisLock {
	return &redisLock{
		redis:   redis,
		lockKey: "",
		ttl:     5,
		wait:    true,
		step:    250,
	}
}

// SetTTl 设置锁时间（秒）初始为5秒
func (redisLock *redisLock) SetTTl(ttl int) *redisLock {
	redisLock.ttl = ttl
	return redisLock
}

// SetWait 是否阻塞等待，不等待的话获取锁失败直接返回false
func (redisLock *redisLock) SetWait(wait bool) *redisLock {
	redisLock.wait = wait
	return redisLock
}

// SetStep 设置尝试获取锁次数，尝试轮询获取锁的次数越多，每次轮询的时间越短
func (redisLock *redisLock) SetStep(step int) *redisLock {
	redisLock.step = step
	return redisLock
}

// SetLockKey 设置锁
func (redisLock *redisLock) SetLockKey(lockKey string) *redisLock {
	redisLock.lockKey = lockKey
	return redisLock
}

// GetLock 获取锁
func (redisLock *redisLock) GetLock() (res bool, err error) {
	// 如果没有key的话直接返回错误
	if redisLock.lockKey == "" {
		return false, errors.New("lock key is required")
	}

	// 随机一个value
	redisLock.token = uuid.NewString()
	index := 0
	times := redisLock.ttl * 1000 / redisLock.step
	for index < times {
		res, err = redisLock.redis.SetNX(context.Background(), redisLock.lockKey, redisLock.token, time.Millisecond*time.Duration(redisLock.ttl*1000)).Result()
		if err != nil {
			return
		}
		if redisLock.wait && !res {
			// milliseconds
			time.Sleep(time.Duration(redisLock.step) * time.Millisecond)
			index++
		} else {
			break
		}
	}

	return
}

// Unlock
// 用EVAL的解锁，删锁带token（锁的值），防止删除锁的时候出现误删
// 结果res一般不需要接收，err接收一下
func (redisLock *redisLock) Unlock() (res bool, err error) {
	res, err = redisLock.redis.Do(context.Background(), "EVAL", "if redis.call(\"GET\", KEYS[1]) == ARGV[1] then return redis.call(\"DEL\", KEYS[1]) else return 0 end", "1", redisLock.lockKey, redisLock.token).Bool()
	return
}
