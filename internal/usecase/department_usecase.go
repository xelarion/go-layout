// Package usecase contains business logic.
package usecase

import (
	"context"

	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/model/gen"
	"github.com/xelarion/go-layout/pkg/errs"
)

// CreateDepartmentParams contains all parameters needed to create a department
type CreateDepartmentParams struct {
	Name        string
	Description string
	Enabled     bool
}

// UpdateDepartmentParams contains all parameters needed to update a department
type UpdateDepartmentParams struct {
	ID          uint
	Name        string
	Description string
	Enabled     bool

	// Fields to track which values are explicitly set
	NameSet        bool
	DescriptionSet bool
	EnabledSet     bool
}

// DepartmentRepository defines methods for department data access.
type DepartmentRepository interface {
	Create(ctx context.Context, department *model.Department) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.Department, int, error)
	FindByID(ctx context.Context, id uint) (*model.Department, error)
	FindByName(ctx context.Context, name string) (*model.Department, error)
	Update(ctx context.Context, department *model.Department) error
	Delete(ctx context.Context, id uint) error
}

// DepartmentUseCase implements business logic for department operations.
type DepartmentUseCase struct {
	repo DepartmentRepository
}

// NewDepartmentUseCase creates a new instance of DepartmentUseCase.
func NewDepartmentUseCase(repo DepartmentRepository) *DepartmentUseCase {
	return &DepartmentUseCase{
		repo: repo,
	}
}

// Create creates a new department.
func (uc *DepartmentUseCase) Create(ctx context.Context, params CreateDepartmentParams) (*model.Department, error) {
	// Check if department already exists
	_, err := uc.repo.FindByName(ctx, params.Name)
	if err != nil {
		if !errs.IsReason(err, errs.ReasonNotFound) {
			return nil, err
		}
	} else {
		return nil, errs.NewBusiness("department name already exists").
			WithReason(errs.ReasonDuplicate)
	}

	// Create department
	department := &model.Department{
		Department: gen.Department{
			Name:        params.Name,
			Description: params.Description,
			Enabled:     params.Enabled,
		},
	}

	if err := uc.repo.Create(ctx, department); err != nil {
		return nil, err
	}

	return department, nil
}

// List returns a list of departments with pagination and filtering.
func (uc *DepartmentUseCase) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.Department, int, error) {
	departments, count, err := uc.repo.List(ctx, filters, limit, offset, sortClause)
	if err != nil {
		return nil, 0, err
	}
	return departments, count, nil
}

// GetByID retrieves a department by ID.
func (uc *DepartmentUseCase) GetByID(ctx context.Context, id uint) (*model.Department, error) {
	department, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return department, nil
}

// Update updates an existing department.
func (uc *DepartmentUseCase) Update(ctx context.Context, params UpdateDepartmentParams) error {
	// Get the existing department
	department, err := uc.repo.FindByID(ctx, params.ID)
	if err != nil {
		return err
	}

	// Update fields that are explicitly set
	if params.NameSet {
		department.Name = params.Name
	}

	if params.DescriptionSet {
		department.Description = params.Description
	}

	if params.EnabledSet {
		department.Enabled = params.Enabled
	}

	if err := uc.repo.Update(ctx, department); err != nil {
		return err
	}
	return nil
}

// Delete removes a department.
func (uc *DepartmentUseCase) Delete(ctx context.Context, id uint) error {
	_, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}
