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

// DepartmentRepository is an implementation of the department repository.
type DepartmentRepository struct {
	db  *gorm.DB
	rds *redis.Client
}

// NewDepartmentRepository creates a new instance of department repository.
func NewDepartmentRepository(db *gorm.DB, rds *redis.Client) *DepartmentRepository {
	return &DepartmentRepository{
		db:  db,
		rds: rds,
	}
}

// Create adds a new department to the database.
func (r *DepartmentRepository) Create(ctx context.Context, department *model.Department) error {
	if err := r.db.WithContext(ctx).Create(department).Error; err != nil {
		return errs.WrapInternal(err, "failed to create department")
	}
	return nil
}

// List retrieves departments with pagination and filtering.
func (r *DepartmentRepository) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.Department, int, error) {
	query := r.db.WithContext(ctx).Model(&model.Department{})

	for field, value := range filters {
		if value != nil {
			query = query.Where(field+" = ?", value)
		}
	}

	var total int64
	if err := query.Model(&model.Department{}).Count(&total).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to count departments")
	}

	if sortClause != "" {
		query = query.Order(sortClause)
	} else {
		query = query.Order("departments.id desc")
	}

	var departments []*model.Department
	if err := query.Limit(limit).Offset(offset).Find(&departments).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to list departments")
	}

	return departments, int(total), nil
}

// FindByID retrieves a department by ID.
func (r *DepartmentRepository) FindByID(ctx context.Context, id uint) (*model.Department, error) {
	var department model.Department
	err := r.db.WithContext(ctx).First(&department, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("department not found").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, errs.WrapInternal(err, "failed to find department by ID")
	}
	return &department, nil
}

func (r *DepartmentRepository) FindByName(ctx context.Context, name string) (*model.Department, error) {
	var department model.Department
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&department).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("department not found").
				WithReason(errs.ReasonNotFound).
				WithMeta("name", name)
		}
		return nil, errs.WrapInternal(err, "failed to find department by name")
	}
	return &department, nil
}

// Update updates a department.
func (r *DepartmentRepository) Update(ctx context.Context, department *model.Department) error {
	result := r.db.WithContext(ctx).Save(department)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "failed to update department")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("department not found").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", department.ID)
	}

	return nil
}

// Delete removes a department by ID.
func (r *DepartmentRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.Department{}, id)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "failed to delete department")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("department not found").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", id)
	}

	return nil
}
