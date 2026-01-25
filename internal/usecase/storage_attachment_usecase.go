// Package usecase contains business logic.
package usecase

import (
	"context"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
)

// StorageAttachmentRepository defines methods for storage_attachment data access.
type StorageAttachmentRepository interface {
	Create(ctx context.Context, storageAttachment *model.StorageAttachment) error
	CreateInBatches(ctx context.Context, storageAttachments []*model.StorageAttachment, batchSize int) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.StorageAttachment, int, error)
	IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error)
	Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error)
	FindByID(ctx context.Context, id uint) (*model.StorageAttachment, error)
	FindByResourceAndName(ctx context.Context, resourceType string, resourceID uint, name enum.StorageAttachmentNameType) (*model.StorageAttachment, error)
	ListByResourceAndName(ctx context.Context, resourceType string, resourceID uint, name enum.StorageAttachmentNameType) ([]*model.StorageAttachment, error)
	ListByResource(ctx context.Context, resourceType string, resourceID uint) ([]*model.StorageAttachment, error)
	Update(ctx context.Context, storageAttachment *model.StorageAttachment, params map[string]any) error
	Delete(ctx context.Context, id uint) error
	BatchDelete(ctx context.Context, ids []uint) error
	DeleteByResource(ctx context.Context, resourceType string, resourceID uint) error
}

// StorageAttachmentUseCase implements business logic for storage_attachment operations.
type StorageAttachmentUseCase struct {
	tx                    Transaction
	storageAttachmentRepo StorageAttachmentRepository
}

// NewStorageAttachmentUseCase creates a new instance of StorageAttachmentUseCase.
func NewStorageAttachmentUseCase(tx Transaction, repo StorageAttachmentRepository) *StorageAttachmentUseCase {
	return &StorageAttachmentUseCase{
		tx:                    tx,
		storageAttachmentRepo: repo,
	}
}
