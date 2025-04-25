// Package repository provides data access implementations.
package repository

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/xelarion/go-layout/internal/infra/cache"
	"github.com/xelarion/go-layout/internal/infra/config"
	"github.com/xelarion/go-layout/internal/infra/database"
	"github.com/xelarion/go-layout/internal/infra/mq"
	"github.com/xelarion/go-layout/internal/usecase"
)

// contextTxKey is the context key for storing tx value
type contextTxKey struct{}

// Data manages data access resources and their lifecycle.
type Data struct {
	db     *gorm.DB      // Database connection
	rds    *redis.Client // Redis client
	mq     *mq.RabbitMQ  // RabbitMQ client
	logger *zap.Logger

	// Closers holds all resources that need to be closed
	closers []func()
}

func NewTransaction(d *Data) usecase.Transaction {
	return d
}

// NewData creates a new Data instance with database, redis, and optional RabbitMQ connections.
// Returns Data instance and a cleanup function.
func NewData(cfgPG *config.PG, cfgRedis *config.Redis, cfgRabbitMQ *config.RabbitMQ, logger *zap.Logger) (*Data, func(), error) {
	data := &Data{
		logger:  logger,
		closers: make([]func(), 0, 3), // Pre-allocate for 3 resources
	}

	// Helper function to handle resource cleanup on error
	cleanup := func() {
		for i := len(data.closers) - 1; i >= 0; i-- {
			data.closers[i]()
		}
	}

	// Initialize database connection
	pg, err := database.NewPostgres(cfgPG, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	data.db = pg.DB
	data.closers = append(data.closers, func() {
		if err := pg.Close(); err != nil {
			logger.Error("Failed to close database",
				zap.String("resource", "postgres"),
				zap.Error(err))
		}
	})

	// Initialize redis connection
	redisClient, err := cache.NewRedis(cfgRedis, logger)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	data.rds = redisClient.Client
	data.closers = append(data.closers, func() {
		if err := redisClient.Close(); err != nil {
			logger.Error("Failed to close redis",
				zap.String("resource", "redis"),
				zap.Error(err))
		}
	})

	// Initialize RabbitMQ if config is provided
	if cfgRabbitMQ != nil {
		rabbitMQ, err := mq.NewRabbitMQ(cfgRabbitMQ, logger)
		if err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
		}
		data.mq = rabbitMQ
		data.closers = append(data.closers, func() {
			rabbitMQ.Close() // RabbitMQ.Close() doesn't return error
		})
	}

	return data, cleanup, nil
}

// Transaction executes a function within a database transaction.
// It automatically commits the transaction if the function returns without error,
// and automatically rolls back the transaction otherwise.
func (d *Data) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create a new context with transaction
		txCtx := context.WithValue(ctx, contextTxKey{}, tx)
		return fn(txCtx)
	})
}

// DB returns the database connection, honoring any transaction in the context.
func (d *Data) DB(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(contextTxKey{}).(*gorm.DB)
	if !ok {
		return d.db.WithContext(ctx)
	}
	return tx
}

// Redis returns the Redis client.
func (d *Data) Redis() *redis.Client {
	return d.rds
}

// RabbitMQ returns the RabbitMQ client if initialized, or nil otherwise.
func (d *Data) RabbitMQ() *mq.RabbitMQ {
	return d.mq
}
