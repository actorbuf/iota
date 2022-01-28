package driver

import (
	"context"
	"time"

	redis2 "github.com/go-redis/redis/v8"
)

const SessionKey = "s:iota:gsession:"

// RedisHandle redis驱动
type RedisHandle struct {
	Driver redis2.UniversalClient
}

func (redis *RedisHandle) getSessionKey(sessionId string) string {
	return SessionKey + sessionId
}

func (redis *RedisHandle) GetConn(c context.Context) *RedisHandle {
	return new(RedisHandle)
}

func (redis *RedisHandle) Close(c context.Context) (bool, error) {
	return true, nil
}

func (redis *RedisHandle) Destroy(c context.Context, sessionId string) (bool, error) {
	_, err := redis.Driver.Del(c, redis.getSessionKey(sessionId)).Result()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (redis *RedisHandle) Gc(c context.Context, maxLifetime int64) (bool, error) {
	return true, nil
}

func (redis *RedisHandle) Open(c context.Context, path string, name string) (bool, error) {
	return true, nil
}

func (redis *RedisHandle) HasSession(c context.Context, sessionId string) (bool, error) {
	has, err := redis.Driver.Exists(c, redis.getSessionKey(sessionId)).Result()
	if err != nil {
		return false, err
	}
	if has == 0 {
		return false, nil
	}
	return true, nil
}

// string版本
func (redis *RedisHandle) Read(c context.Context, sessionId string) (string, error) {
	key := redis.getSessionKey(sessionId)
	readRes, err := redis.Driver.Get(c, key).Result()
	if err != nil && err != redis2.Nil {
		return "", err
	}
	return readRes, nil
}

// string版本
func (redis *RedisHandle) Write(c context.Context, sessionId string, value string, ttl int64) error {
	Key := redis.getSessionKey(sessionId)
	_, err := redis.Driver.Set(c, Key, value, time.Duration(ttl)*time.Second).Result()
	if err != nil {
		return err
	}
	return nil
}
