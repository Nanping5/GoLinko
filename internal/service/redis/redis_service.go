package myredis

import (
	"GoLinko/internal/config"
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var redisClient *redis.Client
var ctx = context.Background()

func init() {
	cfg := config.GetConfig().RedisConfig
	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

}

// SetKeyEx 设置带过期时间的键值对
func SetKeyEx(key string, value any, expiration int) error {
	return redisClient.Set(ctx, key, value, time.Duration(expiration)*time.Second).Err()
}

// GetKey 获取键值对
func GetKey(key string) (string, error) {
	val, err := redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// GetKeyNilIsError 获取键值对，如果键不存在则返回错误
func GetKeyNilIsError(key string) (string, error) {
	return redisClient.Get(ctx, key).Result()
}

// GetKeyWithPrefixNilIsError 根据前缀获取键值对，如果没有匹配的键则返回错误
func GetKeyWithPrefixNilIsError(prefix string) (string, error) {
	keys, err := redisClient.Keys(ctx, prefix+"*").Result()
	if err != nil {
		return "", err
	}
	if len(keys) == 0 {
		return "", redis.Nil
	}
	return redisClient.Get(ctx, keys[0]).Result()
}

// GetKeyWithSuffixNilIsError 根据后缀获取键值对，如果没有匹配的键则返回错误
func GetKeyWithSuffixNilIsError(suffix string) (string, error) {
	keys, err := redisClient.Keys(ctx, "*"+suffix).Result()
	if err != nil {
		return "", err
	}
	if len(keys) == 0 {
		return "", redis.Nil
	}
	return redisClient.Get(ctx, keys[0]).Result()
}

// DelKeyIfExist 删除键，如果键不存在则不删除
func DelKeyIfExist(key string) error {
	return redisClient.Del(ctx, key).Err()
}

// DelKeyWithPrefixIfExist 根据前缀删除键，如果没有匹配的键则不删除
func DelKeyWithPrefixIfExist(prefix string) error {
	keys, err := redisClient.Keys(ctx, prefix+"*").Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return redisClient.Del(ctx, keys...).Err()
	}
	return nil
}

// DelKeyWithSuffixIfExist 根据后缀删除键，如果没有匹配的键则不删除
func DelKeyWithSuffixIfExist(suffix string) error {
	keys, err := redisClient.Keys(ctx, "*"+suffix).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return redisClient.Del(ctx, keys...).Err()
	}
	return nil
}

// DeleteAllRedisKeys 删除所有键
func DeleteAllRedisKeys() error {
	return redisClient.FlushDB(ctx).Err()
}

func DelKeyWithPatternIfExist(pattern string) error {
	keys, err := redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return redisClient.Del(ctx, keys...).Err()
	}
	return nil
}
