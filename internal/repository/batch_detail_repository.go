// Package repository provides data access implementations.
package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/internal/util"
	"github.com/xelarion/go-layout/pkg/errs"
)

var _ usecase.BatchDetailRepository = (*BatchDetailRepository)(nil)

// BatchDetailRepository is an implementation of the batch_detail repository.
type BatchDetailRepository struct {
	data *Data
}

// NewBatchDetailRepository creates a new instance of batch_detail repository.
func NewBatchDetailRepository(data *Data) usecase.BatchDetailRepository {
	return &BatchDetailRepository{
		data: data,
	}
}

// Create adds a new batch_detail to the database.
func (r *BatchDetailRepository) Create(ctx context.Context, batchDetail *model.BatchDetail) error {
	if err := r.data.DB(ctx).Create(batchDetail).Error; err != nil {
		return errs.WrapInternal(err, "failed to create batch_detail")
	}
	return nil
}

// List retrieves batch_details with pagination and filtering.
func (r *BatchDetailRepository) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.BatchDetail, int, error) {
	query := r.data.DB(ctx).Model(&model.BatchDetail{})

	// 处理数据权限的 OR 条件（优先处理）
	if orConditions, ok := filters["_data_scope_or"].([]map[string]any); ok {
		query = ApplyORConditions(query, orConditions)
		delete(filters, "_data_scope_or") // 删除特殊标记，避免后续再次处理
	}

	for field, value := range filters {
		if value == nil {
			continue
		}

		switch field {
		case "key":
			if str, ok := value.(string); ok {
				query = query.Where("batch_details.key LIKE ?", util.EscapeFullLike(str))
			}
		default:
			query = query.Where("batch_details."+field+" = ?", value)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to count batch_details")
	}

	if sortClause != "" {
		query = query.Order(sortClause)
	} else {
		query = query.Order("batch_details.id asc")
	}

	var batchDetails []*model.BatchDetail
	if err := query.Limit(limit).Offset(offset).Find(&batchDetails).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to list batch_details")
	}

	return batchDetails, int(total), nil
}

func (r *BatchDetailRepository) IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error) {
	return IsExists(ctx, r.data.DB(ctx), &model.BatchDetail{}, filters, notFilters)
}

// Count counts the number of batch_details.
func (r *BatchDetailRepository) Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error) {
	return Count(ctx, r.data.DB(ctx), &model.BatchDetail{}, filters, notFilters)
}

// FindByID retrieves a batch_detail by ID.
func (r *BatchDetailRepository) FindByID(ctx context.Context, id uint) (*model.BatchDetail, error) {
	var batchDetail model.BatchDetail
	err := r.data.DB(ctx).First(&batchDetail, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("记录不存在").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, errs.WrapInternal(err, "查询失败")
	}
	return &batchDetail, nil
}

// Update updates a batch_detail.
func (r *BatchDetailRepository) Update(ctx context.Context, batchDetail *model.BatchDetail, params map[string]any) error {
	result := r.data.DB(ctx).Model(batchDetail).Updates(params)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "更新失败")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("记录不存在").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", batchDetail.ID)
	}

	return nil
}

// Delete removes a batch_detail by ID.
func (r *BatchDetailRepository) Delete(ctx context.Context, id uint) error {
	result := r.data.DB(ctx).Delete(&model.BatchDetail{}, id)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "删除失败")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("记录不存在").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", id)
	}

	return nil
}
