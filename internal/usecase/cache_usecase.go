package usecase

import (
	"context"
	"time"
)

// CacheRepository 缓存数据访问接口
// 封装通用的缓存操作，解耦具体的缓存实现（Redis/Memcached/etc）
type CacheRepository interface {
	// Set 设置缓存，ttl=0 表示永不过期
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Get 获取缓存
	Get(ctx context.Context, key string) ([]byte, error)

	// Delete 删除缓存
	Delete(ctx context.Context, key string) error

	// Exists 检查缓存是否存在
	Exists(ctx context.Context, key string) (bool, error)
}
