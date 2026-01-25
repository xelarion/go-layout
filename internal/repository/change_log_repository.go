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

var _ usecase.ChangeLogRepository = (*ChangeLogRepository)(nil)

// ChangeLogRepository is an implementation of the change_log repository.
type ChangeLogRepository struct {
	data *Data
}

// NewChangeLogRepository creates a new instance of change_log repository.
func NewChangeLogRepository(data *Data) usecase.ChangeLogRepository {
	return &ChangeLogRepository{
		data: data,
	}
}

// Create adds a new change_log to the database.
func (r *ChangeLogRepository) Create(ctx context.Context, changeLog *model.ChangeLog) error {
	if err := r.data.DB(ctx).Create(changeLog).Error; err != nil {
		return errs.WrapInternal(err, "创建记录失败")
	}
	return nil
}

// List retrieves change_logs with pagination and filtering.
func (r *ChangeLogRepository) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.ChangeLog, int, error) {
	query := r.data.DB(ctx).Model(&model.ChangeLog{})

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
				query = query.Where("change_logs.id IN (?)", ids)
			}
		case "key":
			if str, ok := value.(string); ok {
				query = query.Where("change_logs.resource_type LIKE @key OR change_logs.resource_id LIKE @key", sql.Named("key", util.EscapeFullLike(str)))
			}
		default:
			query = query.Where("change_logs."+field+" = ?", value)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "查询失败")
	}

	if sortClause != "" {
		query = query.Order(sortClause)
	} else {
		query = query.Order("change_logs.id desc")
	}

	var changeLogs []*model.ChangeLog
	if err := query.Limit(limit).Offset(offset).Preload("Operator", func(db *gorm.DB) *gorm.DB {
		return db.Unscoped()
	}).Find(&changeLogs).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "查询失败")
	}

	return changeLogs, int(total), nil
}

func (r *ChangeLogRepository) IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error) {
	return IsExists(ctx, r.data.DB(ctx), &model.ChangeLog{}, filters, notFilters)
}

// Count counts the number of change_logs.
func (r *ChangeLogRepository) Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error) {
	return Count(ctx, r.data.DB(ctx), &model.ChangeLog{}, filters, notFilters)
}

// FindByID retrieves a change_log by ID.
func (r *ChangeLogRepository) FindByID(ctx context.Context, id uint) (*model.ChangeLog, error) {
	var changeLog model.ChangeLog
	err := r.data.DB(ctx).First(&changeLog, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("记录不存在").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, errs.WrapInternal(err, "查询失败")
	}
	return &changeLog, nil
}

// FindByResourceTypeAndResourceID retrieves a change_log by unique index.
func (r *ChangeLogRepository) FindByResourceTypeAndResourceID(ctx context.Context, resourceType string, resourceID uint) (*model.ChangeLog, error) {
	var changeLog model.ChangeLog
	err := r.data.DB(ctx).Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).First(&changeLog).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("记录不存在").
				WithReason(errs.ReasonNotFound)
		}
		return nil, errs.WrapInternal(err, "查询失败")
	}
	return &changeLog, nil
}

// Update updates a change_log.
func (r *ChangeLogRepository) Update(ctx context.Context, changeLog *model.ChangeLog, params map[string]any) error {
	result := r.data.DB(ctx).Model(changeLog).Updates(params)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "更新失败")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("记录不存在").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", changeLog.ID)
	}

	return nil
}

// Delete removes change_log by ID.
func (r *ChangeLogRepository) Delete(ctx context.Context, id uint) error {
	result := r.data.DB(ctx).Delete(&model.ChangeLog{}, id)
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
