// Package cache provides caching functionality.
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/config"
)

// Redis represents a Redis client instance.
type Redis struct {
	Client *redis.Client
	logger *zap.Logger
}

// NewRedis creates a new Redis client instance.
func NewRedis(cfg *config.Redis, logger *zap.Logger) (*Redis, error) {
	logger = logger.Named("redis")

	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Override pool size if specified
	if cfg.PoolSize > 0 {
		opts.PoolSize = cfg.PoolSize
	}

	client := redis.NewClient(opts)

	// Check connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Connected to Redis",
		zap.String("url", cfg.URL),
		zap.Int("pool_size", cfg.PoolSize))

	return &Redis{
		Client: client,
		logger: logger,
	}, nil
}

// Close closes the Redis client connection.
func (r *Redis) Close() error {
	return r.Client.Close()
}
