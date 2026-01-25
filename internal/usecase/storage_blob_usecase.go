// Package usecase contains business logic.
package usecase

import (
	"context"
	"time"

	"github.com/xelarion/go-layout/internal/model"
)

// StorageBlobRepository defines methods for storage_blob data access.
type StorageBlobRepository interface {
	Create(ctx context.Context, storageBlob *model.StorageBlob) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.StorageBlob, int, error)
	IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error)
	Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error)
	FindByID(ctx context.Context, id uint) (*model.StorageBlob, error)
	FindByKey(ctx context.Context, key string) (*model.StorageBlob, error)
	Update(ctx context.Context, storageBlob *model.StorageBlob, params map[string]any) error
	Delete(ctx context.Context, id uint) error
	// FindNeverAttachedBlobs 查找从未有过 attachment 的 blob（created_at 早于阈值）
	FindNeverAttachedBlobs(ctx context.Context, createdBefore time.Time, limit int) ([]*model.StorageBlob, error)
	// FindDetachedBlobs 查找曾有 attachment 但都已删除超过阈值时间的 blob
	FindDetachedBlobs(ctx context.Context, detachedBefore time.Time, limit int) ([]*model.StorageBlob, error)
}

// StorageBlobUseCase implements business logic for storage_blob operations.
type StorageBlobUseCase struct {
	tx              Transaction
	storageBlobRepo StorageBlobRepository
}

// NewStorageBlobUseCase creates a new instance of StorageBlobUseCase.
func NewStorageBlobUseCase(tx Transaction, repo StorageBlobRepository) *StorageBlobUseCase {
	return &StorageBlobUseCase{
		tx:              tx,
		storageBlobRepo: repo,
	}
}
