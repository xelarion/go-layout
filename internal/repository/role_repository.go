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

var _ usecase.RoleRepository = (*RoleRepository)(nil)

// RoleRepository is an implementation of the role repository.
type RoleRepository struct {
	data *Data
}

// NewRoleRepository creates a new instance of role repository.
func NewRoleRepository(data *Data) usecase.RoleRepository {
	return &RoleRepository{
		data: data,
	}
}

// Create adds a new role to the database.
func (r *RoleRepository) Create(ctx context.Context, role *model.Role) error {
	if err := r.data.DB(ctx).Create(role).Error; err != nil {
		return errs.WrapInternal(err, "failed to create role")
	}
	return nil
}

// List retrieves roles with pagination and filtering.
func (r *RoleRepository) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.Role, int, error) {
	query := r.data.DB(ctx).Model(&model.Role{})

	for field, value := range filters {
		if value == nil {
			continue
		}

		switch field {
		case "name":
			if str, ok := value.(string); ok {
				query = query.Where("roles.name LIKE ?", util.EscapeFullLike(str))
			}
		default:
			query = query.Where("roles."+field+" = ?", value)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to count roles")
	}

	if sortClause != "" {
		query = query.Order(sortClause)
	} else {
		query = query.Order("roles.id desc")
	}

	var roles []*model.Role
	if err := query.Limit(limit).Offset(offset).Find(&roles).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to list roles")
	}

	return roles, int(total), nil
}

func (r *RoleRepository) IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error) {
	return IsExists(ctx, r.data.DB(ctx), &model.Role{}, filters, notFilters)
}

// Count counts the number of roles.
func (r *RoleRepository) Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error) {
	return Count(ctx, r.data.DB(ctx), &model.Role{}, filters, notFilters)
}

// FindByID retrieves a role by ID.
func (r *RoleRepository) FindByID(ctx context.Context, id uint) (*model.Role, error) {
	var role model.Role
	err := r.data.DB(ctx).First(&role, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("role not found").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, errs.WrapInternal(err, "failed to find role by ID")
	}
	return &role, nil
}

// Update updates a role.
func (r *RoleRepository) Update(ctx context.Context, role *model.Role, params map[string]any) error {
	result := r.data.DB(ctx).Model(role).Updates(params)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "failed to update role")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("role not found").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", role.ID)
	}

	return nil
}

// Delete removes a role by ID.
func (r *RoleRepository) Delete(ctx context.Context, id uint) error {
	result := r.data.DB(ctx).Delete(&model.Role{}, id)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "failed to delete role")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("role not found").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", id)
	}

	return nil
}
