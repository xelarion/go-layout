// Package cache provides caching functionality.
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/pkg/config"
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

// Set sets a key-value pair with expiration.
func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := r.Client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		r.logger.Error("Failed to set key in Redis",
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("failed to set key %s in Redis: %w", key, err)
	}
	return nil
}

// Get retrieves a value by key.
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	val, err := r.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Key does not exist
	} else if err != nil {
		r.logger.Error("Failed to get key from Redis",
			zap.String("key", key),
			zap.Error(err))
		return "", fmt.Errorf("failed to get key %s from Redis: %w", key, err)
	}
	return val, nil
}

// Delete removes a key.
func (r *Redis) Delete(ctx context.Context, key string) error {
	err := r.Client.Del(ctx, key).Err()
	if err != nil {
		r.logger.Error("Failed to delete key from Redis",
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("failed to delete key %s from Redis: %w", key, err)
	}
	return nil
}

// Close closes the Redis client connection.
func (r *Redis) Close() error {
	return r.Client.Close()
}
