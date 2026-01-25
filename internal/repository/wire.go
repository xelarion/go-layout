// Package repository provides data access implementations.
package repository

import (
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet provides all repository layer dependencies.
var ProviderSet = wire.NewSet(
	NewData,
	NewTransaction,
	ProvideRedis,
	ProvideRedsync,
	ProvideRabbitMQExchange,

	// Repositories sorted by name
	NewAmqpMessageRepository,
	NewAnnouncementRepository,
	NewBatchDetailRepository,
	NewBatchRecordRepository,
	NewCacheRepository,
	NewChangeLogDetailRepository,
	NewChangeLogRepository,
	NewCodeRuleRepository,
	NewCodeSequenceRepository,
	NewExportTaskRepository,
	NewMQRepository,
	NewNotificationRepository,
	NewOrganizationRepository,
	NewPositionRepository,
	NewRoleRepository,
	NewStorageAttachmentRepository,
	NewStorageBlobRepository,
	NewUserNotificationRepository,
	NewUserRepository,
)

// ProvideRedis provides Redis client from Data
func ProvideRedis(data *Data) *redis.Client {
	return data.Redis()
}

// ProvideRedsync provides Redsync instance for distributed locking
func ProvideRedsync(rds *redis.Client) *redsync.Redsync {
	pool := goredis.NewPool(rds)
	return redsync.New(pool)
}

// ProvideRabbitMQExchange provides RabbitMQ exchange name from Data
func ProvideRabbitMQExchange(data *Data) string {
	return data.RabbitMQExchange()
}
