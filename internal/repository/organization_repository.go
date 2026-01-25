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

var _ usecase.OrganizationRepository = (*OrganizationRepository)(nil)

// OrganizationRepository is an implementation of the organization repository.
type OrganizationRepository struct {
	data *Data
}

// NewOrganizationRepository creates a new instance of organization repository.
func NewOrganizationRepository(data *Data) usecase.OrganizationRepository {
	return &OrganizationRepository{
		data: data,
	}
}

// Create adds a new organization to the database.
func (r *OrganizationRepository) Create(ctx context.Context, organization *model.Organization) error {
	if err := r.data.DB(ctx).Create(organization).Error; err != nil {
		return errs.WrapInternal(err, "failed to create organization")
	}
	return nil
}

// List retrieves organizations with pagination and filtering.
func (r *OrganizationRepository) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.Organization, int, error) {
	query := r.data.DB(ctx).Model(&model.Organization{})

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
				query = query.Where("organizations.name LIKE @key OR code LIKE @key", sql.Named("key", util.EscapeFullLike(str)))
			}
		default:
			query = query.Where("organizations."+field+" = ?", value)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to count organizations")
	}

	if sortClause != "" {
		query = query.Order(sortClause)
	} else {
		query = query.Order("organizations.id desc")
	}

	var organizations []*model.Organization
	if err := query.Limit(limit).Offset(offset).Find(&organizations).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to list organizations")
	}

	return organizations, int(total), nil
}

func (r *OrganizationRepository) IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error) {
	return IsExists(ctx, r.data.DB(ctx), &model.Organization{}, filters, notFilters)
}

// Count counts the number of organizations.
func (r *OrganizationRepository) Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error) {
	return Count(ctx, r.data.DB(ctx), &model.Organization{}, filters, notFilters)
}

// FindByID retrieves a organization by ID.
func (r *OrganizationRepository) FindByID(ctx context.Context, id uint) (*model.Organization, error) {
	var organization model.Organization
	err := r.data.DB(ctx).First(&organization, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("organization not found").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, errs.WrapInternal(err, "failed to find organization by ID")
	}
	return &organization, nil
}

// Update updates a organization.
func (r *OrganizationRepository) Update(ctx context.Context, organization *model.Organization, params map[string]any) error {
	result := r.data.DB(ctx).Model(organization).Updates(params)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "failed to update organization")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("organization not found").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", organization.ID)
	}

	return nil
}

// Delete removes a organization by ID.
func (r *OrganizationRepository) Delete(ctx context.Context, id uint) error {
	result := r.data.DB(ctx).Delete(&model.Organization{}, id)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "failed to delete organization")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("organization not found").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", id)
	}

	return nil
}
