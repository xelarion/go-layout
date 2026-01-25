// Package repository provides data access implementations.
package repository

import (
	"context"
	"database/sql"
	"errors"

	"gorm.io/gorm"

	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/internal/util"
	"github.com/xelarion/go-layout/pkg/errs"
)

var _ usecase.ChangeLogDetailRepository = (*ChangeLogDetailRepository)(nil)

// ChangeLogDetailRepository is an implementation of the change_log_detail repository.
type ChangeLogDetailRepository struct {
	data *Data
}

// NewChangeLogDetailRepository creates a new instance of change_log_detail repository.
func NewChangeLogDetailRepository(data *Data) usecase.ChangeLogDetailRepository {
	return &ChangeLogDetailRepository{
		data: data,
	}
}

// Create adds a new change_log_detail to the database.
func (r *ChangeLogDetailRepository) Create(ctx context.Context, changeLogDetail *model.ChangeLogDetail) error {
	if err := r.data.DB(ctx).Create(changeLogDetail).Error; err != nil {
		return errs.WrapInternal(err, "创建记录失败")
	}
	return nil
}

// CreateInBatches 批量创建变更记录详情
func (r *ChangeLogDetailRepository) CreateInBatches(ctx context.Context, details []*model.ChangeLogDetail, batchSize int) error {
	if err := r.data.DB(ctx).CreateInBatches(details, batchSize).Error; err != nil {
		return errs.WrapInternal(err, "批量创建记录失败")
	}
	return nil
}

// List retrieves change_log_details with pagination and filtering.
func (r *ChangeLogDetailRepository) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.ChangeLogDetail, int, error) {
	query := r.data.DB(ctx).Model(&model.ChangeLogDetail{})

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
		case "ids":
			if ids, ok := value.([]uint); ok {
				query = query.Where("change_log_details.id IN (?)", ids)
			}
		case "key":
			if str, ok := value.(string); ok {
				query = query.Where("", sql.Named("key", util.EscapeFullLike(str)))
			}
		default:
			query = query.Where("change_log_details."+field+" = ?", value)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "查询失败")
	}

	if sortClause != "" {
		query = query.Order(sortClause)
	} else {
		query = query.Order("change_log_details.id desc")
	}

	var changeLogDetails []*model.ChangeLogDetail
	if err := query.Limit(limit).Offset(offset).Find(&changeLogDetails).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "查询失败")
	}

	return changeLogDetails, int(total), nil
}

func (r *ChangeLogDetailRepository) IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error) {
	return IsExists(ctx, r.data.DB(ctx), &model.ChangeLogDetail{}, filters, notFilters)
}

// Count counts the number of change_log_details.
func (r *ChangeLogDetailRepository) Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error) {
	return Count(ctx, r.data.DB(ctx), &model.ChangeLogDetail{}, filters, notFilters)
}

// FindByID retrieves a change_log_detail by ID.
func (r *ChangeLogDetailRepository) FindByID(ctx context.Context, id uint) (*model.ChangeLogDetail, error) {
	var changeLogDetail model.ChangeLogDetail
	err := r.data.DB(ctx).First(&changeLogDetail, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("记录不存在").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, errs.WrapInternal(err, "查询失败")
	}
	return &changeLogDetail, nil
}

// Update updates a change_log_detail.
func (r *ChangeLogDetailRepository) Update(ctx context.Context, changeLogDetail *model.ChangeLogDetail, params map[string]any) error {
	result := r.data.DB(ctx).Model(changeLogDetail).Updates(params)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "更新失败")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("记录不存在").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", changeLogDetail.ID)
	}

	return nil
}

// Delete removes change_log_detail by ID.
func (r *ChangeLogDetailRepository) Delete(ctx context.Context, id uint) error {
	result := r.data.DB(ctx).Delete(&model.ChangeLogDetail{}, id)
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
