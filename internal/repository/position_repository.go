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

var _ usecase.PositionRepository = (*PositionRepository)(nil)

// PositionRepository is an implementation of the position repository.
type PositionRepository struct {
	data *Data
}

// NewPositionRepository creates a new instance of position repository.
func NewPositionRepository(data *Data) usecase.PositionRepository {
	return &PositionRepository{
		data: data,
	}
}

// Create adds a new position to the database.
func (r *PositionRepository) Create(ctx context.Context, position *model.Position) error {
	if err := r.data.DB(ctx).Create(position).Error; err != nil {
		return errs.WrapInternal(err, "failed to create position")
	}
	return nil
}

// List retrieves positions with pagination and filtering.
func (r *PositionRepository) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.Position, int, error) {
	query := r.data.DB(ctx).Model(&model.Position{})

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
				query = query.Where("positions.name LIKE @key OR code LIKE @key", sql.Named("key", util.EscapeFullLike(str)))
			}
		default:
			query = query.Where("positions."+field+" = ?", value)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to count positions")
	}

	if sortClause != "" {
		query = query.Order(sortClause)
	} else {
		query = query.Order("positions.id desc")
	}

	var positions []*model.Position
	if err := query.Limit(limit).Offset(offset).Find(&positions).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to list positions")
	}

	return positions, int(total), nil
}

func (r *PositionRepository) IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error) {
	return IsExists(ctx, r.data.DB(ctx), &model.Position{}, filters, notFilters)
}

// Count counts the number of positions.
func (r *PositionRepository) Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error) {
	return Count(ctx, r.data.DB(ctx), &model.Position{}, filters, notFilters)
}

// FindByID retrieves a position by ID.
func (r *PositionRepository) FindByID(ctx context.Context, id uint) (*model.Position, error) {
	var position model.Position
	err := r.data.DB(ctx).First(&position, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("position not found").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, errs.WrapInternal(err, "failed to find position by ID")
	}
	return &position, nil
}

// Update updates a position.
func (r *PositionRepository) Update(ctx context.Context, position *model.Position, params map[string]any) error {
	result := r.data.DB(ctx).Model(position).Updates(params)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "failed to update position")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("position not found").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", position.ID)
	}

	return nil
}

// Delete removes a position by ID.
func (r *PositionRepository) Delete(ctx context.Context, id uint) error {
	result := r.data.DB(ctx).Delete(&model.Position{}, id)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "failed to delete position")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("position not found").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", id)
	}

	return nil
}
