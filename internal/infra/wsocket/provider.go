package wsocket

import (
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ManagerProvider 提供 Manager 实例并启动事件循环
func ManagerProvider(logger *zap.Logger) *Manager {
	m := NewManager(logger)
	go m.Run()
	return m
}

// BroadcasterProvider 提供 Broadcaster 实例（web-api 使用，带订阅功能）
func BroadcasterProvider(redisClient *redis.Client, manager *Manager, logger *zap.Logger) Broadcaster {
	return NewRedisBroadcaster(redisClient, logger, WithSubscriber(manager))
}

// PublisherOnlyBroadcasterProvider 提供仅发布的 Broadcaster（task 服务使用）
func PublisherOnlyBroadcasterProvider(redisClient *redis.Client, logger *zap.Logger) Broadcaster {
	return NewRedisBroadcaster(redisClient, logger)
}
