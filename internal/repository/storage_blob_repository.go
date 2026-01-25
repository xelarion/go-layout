// Package repository provides data access implementations.
package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/errs"
)

var _ usecase.StorageBlobRepository = (*StorageBlobRepository)(nil)

// StorageBlobRepository is an implementation of the storage_blob repository.
type StorageBlobRepository struct {
	data *Data
}

// NewStorageBlobRepository creates a new instance of storage_blob repository.
func NewStorageBlobRepository(data *Data) usecase.StorageBlobRepository {
	return &StorageBlobRepository{
		data: data,
	}
}

// Create adds a new storage_blob to the database.
func (r *StorageBlobRepository) Create(ctx context.Context, storageBlob *model.StorageBlob) error {
	if err := r.data.DB(ctx).Create(storageBlob).Error; err != nil {
		if IsDuplicateKeyError(err) {
			return errs.NewBusiness("storage_blob already exists").
				WithReason(errs.ReasonDuplicate).
				WithMeta("key", storageBlob.Key)
		}

		return errs.WrapInternal(err, "failed to create storage_blob")
	}
	return nil
}

// List retrieves storage_blobs with pagination and filtering.
func (r *StorageBlobRepository) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.StorageBlob, int, error) {
	query := r.data.DB(ctx).Model(&model.StorageBlob{})

	for field, value := range filters {
		if value == nil {
			continue
		}

		switch field {
		default:
			query = query.Where("storage_blobs."+field+" = ?", value)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to count storage_blobs")
	}

	if sortClause != "" {
		query = query.Order(sortClause)
	} else {
		query = query.Order("storage_blobs.id desc")
	}

	var storageBlobs []*model.StorageBlob
	if err := query.Limit(limit).Offset(offset).Find(&storageBlobs).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to list storage_blobs")
	}

	return storageBlobs, int(total), nil
}

// IsExists checks if a storage_blob exists in the database.
func (r *StorageBlobRepository) IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error) {
	return IsExists(ctx, r.data.DB(ctx), &model.StorageBlob{}, filters, notFilters)
}

// Count counts the number of storage_blobs.
func (r *StorageBlobRepository) Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error) {
	return Count(ctx, r.data.DB(ctx), &model.StorageBlob{}, filters, notFilters)
}

// FindByID retrieves a storage_blob by ID.
func (r *StorageBlobRepository) FindByID(ctx context.Context, id uint) (*model.StorageBlob, error) {
	var storageBlob model.StorageBlob
	err := r.data.DB(ctx).First(&storageBlob, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("storage_blob not found").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, errs.WrapInternal(err, "failed to find storage_blob by ID")
	}
	return &storageBlob, nil
}

// FindByKey retrieves a storage_blob by Key.
func (r *StorageBlobRepository) FindByKey(ctx context.Context, key string) (*model.StorageBlob, error) {
	var storageBlob model.StorageBlob
	err := r.data.DB(ctx).Where("key = ?", key).First(&storageBlob).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("storage_blob not found").
				WithReason(errs.ReasonNotFound).
				WithMeta("key", key)
		}
		return nil, errs.WrapInternal(err, "failed to find storage_blob by Key")
	}
	return &storageBlob, nil
}

// Update updates a storage_blob.
func (r *StorageBlobRepository) Update(ctx context.Context, storageBlob *model.StorageBlob, params map[string]any) error {
	result := r.data.DB(ctx).Model(storageBlob).Updates(params)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "failed to update storage_blob")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("storage_blob not found").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", storageBlob.ID)
	}

	return nil
}

// Delete removes a storage_blob by ID.
func (r *StorageBlobRepository) Delete(ctx context.Context, id uint) error {
	result := r.data.DB(ctx).Delete(&model.StorageBlob{}, id)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "failed to delete storage_blob")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("storage_blob not found").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", id)
	}

	return nil
}

// FindNeverAttachedBlobs 查找从未有过 attachment 的 blob
// 使用 NOT EXISTS 子查询，效率高且语义清晰
func (r *StorageBlobRepository) FindNeverAttachedBlobs(ctx context.Context, createdBefore time.Time, limit int) ([]*model.StorageBlob, error) {
	var blobs []*model.StorageBlob
	err := r.data.DB(ctx).
		Model(&model.StorageBlob{}).
		Where("NOT EXISTS (SELECT 1 FROM storage_attachments WHERE storage_attachments.storage_blob_id = storage_blobs.id)").
		Where("created_at < ?", createdBefore).
		Limit(limit).
		Find(&blobs).
		Error
	if err != nil {
		return nil, errs.WrapInternal(err, "failed to find never attached blobs")
	}
	return blobs, nil
}

// FindDetachedBlobs 查找曾有 attachment 但都已删除超过阈值时间的 blob
// 条件：所有 attachment 都已删除（deleted_at > 0），且最后删除时间早于阈值
func (r *StorageBlobRepository) FindDetachedBlobs(ctx context.Context, detachedBefore time.Time, limit int) ([]*model.StorageBlob, error) {
	var blobs []*model.StorageBlob
	detachedBeforeNano := detachedBefore.UnixNano()

	// PostgreSQL: 使用 EXISTS + BOOL_AND
	err := r.data.DB(ctx).
		Model(&model.StorageBlob{}).
		Where(`EXISTS (
			SELECT 1 FROM storage_attachments
			WHERE storage_attachments.storage_blob_id = storage_blobs.id
			GROUP BY storage_attachments.storage_blob_id
			HAVING BOOL_AND(storage_attachments.deleted_at <> 0) AND MAX(storage_attachments.deleted_at) < ?
		)`, detachedBeforeNano).
		Limit(limit).
		Find(&blobs).
		Error

	// MySQL: 使用 IN + MIN/MAX（如需切换数据库，使用以下写法）
	// subQuery := r.data.DB(ctx).
	// 	Table("storage_attachments").
	// 	Select("storage_blob_id").
	// 	Group("storage_blob_id").
	// 	Having("MIN(deleted_at) > 0").
	// 	Having("MAX(deleted_at) < ?", detachedBeforeNano)
	// err := r.data.DB(ctx).
	// 	Model(&model.StorageBlob{}).
	// 	Where("id IN (?)", subQuery).
	// 	Limit(limit).
	// 	Find(&blobs).
	// 	Error

	if err != nil {
		return nil, errs.WrapInternal(err, "failed to find detached blobs")
	}
	return blobs, nil
}
