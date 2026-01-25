// Package usecase contains business logic.
package usecase

import (
	"context"

	"github.com/xelarion/go-layout/internal/model"
)

// ChangeLogDetailRepository defines methods for change_log_detail data access.
type ChangeLogDetailRepository interface {
	Create(ctx context.Context, changeLogDetail *model.ChangeLogDetail) error
	CreateInBatches(ctx context.Context, details []*model.ChangeLogDetail, batchSize int) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.ChangeLogDetail, int, error)
	IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error)
	Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error)
	FindByID(ctx context.Context, id uint) (*model.ChangeLogDetail, error)
	Update(ctx context.Context, changeLogDetail *model.ChangeLogDetail, params map[string]any) error
	Delete(ctx context.Context, id uint) error
}

// ChangeLogDetailUseCase implements business logic for change_log_detail operations.
type ChangeLogDetailUseCase struct {
	tx                  Transaction
	changeLogDetailRepo ChangeLogDetailRepository
}

// NewChangeLogDetailUseCase creates a new instance of ChangeLogDetailUseCase.
func NewChangeLogDetailUseCase(
	tx Transaction,
	repo ChangeLogDetailRepository,
) *ChangeLogDetailUseCase {
	return &ChangeLogDetailUseCase{
		tx:                  tx,
		changeLogDetailRepo: repo,
	}
}
