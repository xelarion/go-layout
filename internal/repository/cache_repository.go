package repository

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/errs"
)

// CacheRepository 缓存仓储实现
type CacheRepository struct {
	rds *redis.Client
}

// NewCacheRepository 创建缓存仓储实例
func NewCacheRepository(rds *redis.Client) usecase.CacheRepository {
	return &CacheRepository{
		rds: rds,
	}
}

// Set 设置缓存
func (r *CacheRepository) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	err := r.rds.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return errs.WrapInternal(err, "failed to set cache")
	}
	return nil
}

// Get 获取缓存
func (r *CacheRepository) Get(ctx context.Context, key string) ([]byte, error) {
	value, err := r.rds.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errs.NewBusiness("cache key not found").
				WithReason(errs.ReasonNotFound).
				WithMeta("key", key)
		}
		return nil, errs.WrapInternal(err, "failed to get cache")
	}
	return value, nil
}

// Delete 删除缓存
func (r *CacheRepository) Delete(ctx context.Context, key string) error {
	err := r.rds.Del(ctx, key).Err()
	if err != nil {
		return errs.WrapInternal(err, "failed to delete cache")
	}
	return nil
}

// Exists 检查缓存是否存在
func (r *CacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.rds.Exists(ctx, key).Result()
	if err != nil {
		return false, errs.WrapInternal(err, "failed to check cache existence")
	}
	return count > 0, nil
}
