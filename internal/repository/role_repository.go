// Package repository provides data access implementations.
package repository

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/pkg/errs"
)

// RoleRepository is an implementation of the role repository.
type RoleRepository struct {
	db  *gorm.DB
	rds *redis.Client
}

// NewRoleRepository creates a new instance of role repository.
func NewRoleRepository(db *gorm.DB, rds *redis.Client) *RoleRepository {
	return &RoleRepository{
		db:  db,
		rds: rds,
	}
}

// Create adds a new role to the database.
func (r *RoleRepository) Create(ctx context.Context, role *model.Role) error {
	if err := r.db.WithContext(ctx).Create(role).Error; err != nil {
		return errs.WrapInternal(err, "failed to create role")
	}
	return nil
}

// List retrieves roles with pagination and filtering.
func (r *RoleRepository) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.Role, int, error) {
	query := r.db.WithContext(ctx).Model(&model.Role{})

	for field, value := range filters {
		if value == nil {
			continue
		}

		switch field {
		case "name":
			if str, ok := value.(string); ok {
				query = query.Where(field+" LIKE ?", "%"+str+"%")
			}
		default:
			query = query.Where(field+" = ?", value)
		}
	}

	var total int64
	if err := query.Model(&model.Role{}).Count(&total).Error; err != nil {
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
	return IsExists(ctx, r.db, &model.Role{}, filters, notFilters)
}

// FindByID retrieves a role by ID.
func (r *RoleRepository) FindByID(ctx context.Context, id uint) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).First(&role, id).Error
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
func (r *RoleRepository) Update(ctx context.Context, role *model.Role) error {
	result := r.db.WithContext(ctx).Save(role)
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
	result := r.db.WithContext(ctx).Delete(&model.Role{}, id)
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

// CountUsersByRoleID counts the number of users in a role.
func (r *RoleRepository) CountUsersByRoleID(ctx context.Context, roleID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("role_id = ?", roleID).
		Where("enabled = ?", true).
		Count(&count).Error

	if err != nil {
		return 0, errs.WrapInternal(err, "failed to count users by role ID")
	}
	return count, nil
}
