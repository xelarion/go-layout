// Package repository provides data access implementations.
package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/errs"
)

var _ usecase.StorageAttachmentRepository = (*StorageAttachmentRepository)(nil)

// StorageAttachmentRepository is an implementation of the storage_attachment repository.
type StorageAttachmentRepository struct {
	data *Data
}

// NewStorageAttachmentRepository creates a new instance of storage_attachment repository.
func NewStorageAttachmentRepository(data *Data) usecase.StorageAttachmentRepository {
	return &StorageAttachmentRepository{
		data: data,
	}
}

// Create adds a new storage_attachment to the database.
func (r *StorageAttachmentRepository) Create(ctx context.Context, storageAttachment *model.StorageAttachment) error {
	if err := r.data.DB(ctx).Create(storageAttachment).Error; err != nil {
		return errs.WrapInternal(err, "failed to create storage_attachment")
	}
	return nil
}

func (r *StorageAttachmentRepository) CreateInBatches(ctx context.Context, storageAttachments []*model.StorageAttachment, batchSize int) error {
	if err := r.data.DB(ctx).CreateInBatches(storageAttachments, batchSize).Error; err != nil {
		return errs.WrapInternal(err, "failed to create storage_attachments in batches")
	}
	return nil
}

// List retrieves storage_attachments with pagination and filtering.
func (r *StorageAttachmentRepository) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.StorageAttachment, int, error) {
	query := r.data.DB(ctx).Model(&model.StorageAttachment{})

	for field, value := range filters {
		if value == nil {
			continue
		}

		switch field {
		default:
			query = query.Where("storage_attachments."+field+" = ?", value)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to count storage_attachments")
	}

	if sortClause != "" {
		query = query.Order(sortClause)
	} else {
		query = query.Order("storage_attachments.id desc")
	}

	var storageAttachments []*model.StorageAttachment
	if err := query.Limit(limit).Offset(offset).Preload("StorageBlob").Find(&storageAttachments).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to list storage_attachments")
	}

	return storageAttachments, int(total), nil
}

// IsExists checks if a storage_attachment exists in the database.
func (r *StorageAttachmentRepository) IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error) {
	return IsExists(ctx, r.data.DB(ctx), &model.StorageAttachment{}, filters, notFilters)
}

// Count counts the number of storage_attachments.
func (r *StorageAttachmentRepository) Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error) {
	return Count(ctx, r.data.DB(ctx), &model.StorageAttachment{}, filters, notFilters)
}

// FindByID retrieves a storage_attachment by ID.
func (r *StorageAttachmentRepository) FindByID(ctx context.Context, id uint) (*model.StorageAttachment, error) {
	var storageAttachment model.StorageAttachment
	err := r.data.DB(ctx).
		Preload("StorageBlob").
		First(&storageAttachment, id).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("storage_attachment not found").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, errs.WrapInternal(err, "failed to find storage_attachment by ID")
	}
	return &storageAttachment, nil
}

func (r *StorageAttachmentRepository) FindByResourceAndName(ctx context.Context, resourceType string, resourceID uint, name enum.StorageAttachmentNameType) (*model.StorageAttachment, error) {
	var storageAttachment model.StorageAttachment
	err := r.data.DB(ctx).
		Where("resource_type = ? AND resource_id = ? AND name = ?", resourceType, resourceID, name).
		Preload("StorageBlob").
		First(&storageAttachment).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("storage_attachment not found").
				WithReason(errs.ReasonNotFound).
				WithMeta("resource_type", resourceType).
				WithMeta("resource_id", resourceID).
				WithMeta("name", name)
		}
		return nil, errs.WrapInternal(err, "failed to find storage_attachment by resource and name")
	}
	return &storageAttachment, nil
}

func (r *StorageAttachmentRepository) ListByResourceAndName(ctx context.Context, resourceType string, resourceID uint, name enum.StorageAttachmentNameType) ([]*model.StorageAttachment, error) {
	var storageAttachments []*model.StorageAttachment
	err := r.data.DB(ctx).
		Where("resource_type = ? AND resource_id = ? AND name = ?", resourceType, resourceID, name).
		Preload("StorageBlob").
		Find(&storageAttachments).
		Error
	if err != nil {
		return nil, errs.WrapInternal(err, "failed to list storage_attachments by resource and name")
	}
	return storageAttachments, nil
}

func (r *StorageAttachmentRepository) ListByResource(ctx context.Context, resourceType string, resourceID uint) ([]*model.StorageAttachment, error) {
	var storageAttachments []*model.StorageAttachment
	err := r.data.DB(ctx).
		Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).
		Preload("StorageBlob").
		Find(&storageAttachments).
		Error
	if err != nil {
		return nil, errs.WrapInternal(err, "failed to list storage_attachments by resource")
	}
	return storageAttachments, nil
}

// Update updates a storage_attachment.
func (r *StorageAttachmentRepository) Update(ctx context.Context, storageAttachment *model.StorageAttachment, params map[string]any) error {
	result := r.data.DB(ctx).Model(storageAttachment).Updates(params)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "failed to update storage_attachment")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("storage_attachment not found").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", storageAttachment.ID)
	}

	return nil
}

// Delete removes a storage_attachment by ID.
func (r *StorageAttachmentRepository) Delete(ctx context.Context, id uint) error {
	result := r.data.DB(ctx).Delete(&model.StorageAttachment{}, id)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "failed to delete storage_attachment")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("storage_attachment not found").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", id)
	}

	return nil
}

func (r *StorageAttachmentRepository) BatchDelete(ctx context.Context, ids []uint) error {
	result := r.data.DB(ctx).Delete(&model.StorageAttachment{}, ids)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "failed to delete storage_attachments")
	}

	if result.RowsAffected != int64(len(ids)) {
		return errs.NewBusiness("failed to delete storage_attachments").
			WithReason(errs.ReasonNotFound).
			WithMeta("ids", ids)
	}

	return nil
}

func (r *StorageAttachmentRepository) DeleteByResource(ctx context.Context, resourceType string, resourceID uint) error {
	result := r.data.DB(ctx).
		Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).
		Delete(&model.StorageAttachment{})
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "failed to delete storage_attachments by resource")
	}
	return nil
}
