package myredis

import (
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	// UserOnlineKeyPrefix 用户在线状态 Key 前缀
	UserOnlineKeyPrefix = "user:online:"
	// UserOnlineTTL 用户在线状态过期时间
	UserOnlineTTL = 300 // 5分钟
	// UserHeartbeatInterval 心跳续期间隔
	UserHeartbeatInterval = 60 // 1分钟
)

// SetUserOnline 设置用户在线状态
// userID: 用户ID
// instanceID: 当前实例ID
func SetUserOnline(userID, instanceID string) error {
	key := UserOnlineKeyPrefix + userID
	return redisClient.Set(ctx, key, instanceID, UserOnlineTTL*time.Second).Err()
}

// GetUserInstance 获取用户所在的实例ID
// 返回空字符串表示用户不在线
func GetUserInstance(userID string) (string, error) {
	key := UserOnlineKeyPrefix + userID
	val, err := redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // 用户不在线
	}
	return val, err
}

// RemoveUserOnline 移除用户在线状态（用户下线时调用）
func RemoveUserOnline(userID string) error {
	key := UserOnlineKeyPrefix + userID
	return redisClient.Del(ctx, key).Err()
}

// RefreshUserOnline 刷新用户在线状态（心跳续期）
func RefreshUserOnline(userID string) error {
	key := UserOnlineKeyPrefix + userID
	return redisClient.Expire(ctx, key, UserOnlineTTL*time.Second).Err()
}

// IsUserOnline 检查用户是否在线
func IsUserOnline(userID string) (bool, error) {
	key := UserOnlineKeyPrefix + userID
	val, err := redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

// SetUserOnlineWithTTL 设置用户在线状态（自定义过期时间）
func SetUserOnlineWithTTL(userID, instanceID string, ttl time.Duration) error {
	key := UserOnlineKeyPrefix + userID
	return redisClient.Set(ctx, key, instanceID, ttl).Err()
}
