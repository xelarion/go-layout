package tasks

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/task"
	"github.com/xelarion/go-layout/internal/task/scheduler"
)

// StorageCleanerHandler 清理孤立存储文件的定时任务
type StorageCleanerHandler struct {
	deps   *task.Dependencies
	logger *zap.Logger
}

// NewStorageCleanerHandler 创建存储清理任务实例
func NewStorageCleanerHandler(deps *task.Dependencies, logger *zap.Logger) *StorageCleanerHandler {
	return &StorageCleanerHandler{
		deps:   deps,
		logger: logger.Named("storage-cleaner-handler"),
	}
}

// Register 注册定时任务
func (t *StorageCleanerHandler) Register(s *scheduler.Scheduler) error {
	// 每天凌晨 3 点执行清理任务
	return s.Register("storage-cleaner", "0 0 3 * * *", t.Execute)
}

// Execute 执行清理任务
func (t *StorageCleanerHandler) Execute(ctx context.Context) error {
	batchSize := 100

	// 1. 清理从未有过 attachment 的 blob
	if err := t.cleanNeverAttachedBlobs(ctx, time.Now().Add(-48*time.Hour), batchSize); err != nil {
		return err
	}

	// 2. 清理曾有 attachment 但都已删除超过阈值时间的 blob
	if err := t.cleanDetachedBlobs(ctx, time.Now().Add(-7*24*time.Hour), batchSize); err != nil {
		return err
	}

	return nil
}

// cleanNeverAttachedBlobs 清理从未有过 attachment 的 blob
func (t *StorageCleanerHandler) cleanNeverAttachedBlobs(ctx context.Context, createdBefore time.Time, batchSize int) error {
	for {
		blobs, err := t.deps.StorageBlobRepo.FindNeverAttachedBlobs(ctx, createdBefore, batchSize)
		if err != nil {
			return err
		}

		if len(blobs) == 0 {
			break
		}

		for _, blob := range blobs {
			if err := t.deps.StorageUseCase.DeleteBlob(ctx, blob.ID, blob.Key); err != nil {
				t.logger.Warn("failed to delete never attached blob",
					zap.Uint("blob_id", blob.ID),
					zap.String("key", blob.Key),
					zap.Error(err),
				)
			}
		}

		if len(blobs) < batchSize {
			break
		}
	}
	return nil
}

// cleanDetachedBlobs 清理曾有 attachment 但都已删除超过阈值时间的 blob
func (t *StorageCleanerHandler) cleanDetachedBlobs(ctx context.Context, detachedBefore time.Time, batchSize int) error {
	for {
		blobs, err := t.deps.StorageBlobRepo.FindDetachedBlobs(ctx, detachedBefore, batchSize)
		if err != nil {
			return err
		}

		if len(blobs) == 0 {
			break
		}

		for _, blob := range blobs {
			if err := t.deps.StorageUseCase.DeleteBlob(ctx, blob.ID, blob.Key); err != nil {
				t.logger.Warn("failed to delete detached blob",
					zap.Uint("blob_id", blob.ID),
					zap.String("key", blob.Key),
					zap.Error(err),
				)
			}
		}

		if len(blobs) < batchSize {
			break
		}
	}
	return nil
}
