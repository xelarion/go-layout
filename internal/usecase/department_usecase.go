// Package usecase contains business logic.
package usecase

import (
	"context"

	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/model/gen"
	"github.com/xelarion/go-layout/pkg/errs"
)

// Department contains department data
type Department struct {
	*model.Department
	UserCount int64
}

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
	IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error)
	Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error)
	FindByID(ctx context.Context, id uint) (*model.Department, error)
	Update(ctx context.Context, department *model.Department) error
	Delete(ctx context.Context, id uint) error
}

// DepartmentUseCase implements business logic for department operations.
type DepartmentUseCase struct {
	departmentRepo DepartmentRepository
	userRepo       UserRepository
}

// NewDepartmentUseCase creates a new instance of DepartmentUseCase.
func NewDepartmentUseCase(repo DepartmentRepository, userRepo UserRepository) *DepartmentUseCase {
	return &DepartmentUseCase{
		departmentRepo: repo,
		userRepo:       userRepo,
	}
}

// Create creates a new department.
func (uc *DepartmentUseCase) Create(ctx context.Context, params CreateDepartmentParams) (uint, error) {
	// Check if department already exists
	exists, err := uc.departmentRepo.IsExists(ctx, map[string]any{"name": params.Name}, nil)
	if err != nil {
		return 0, err
	}

	if exists {
		return 0, errs.NewBusiness("department name already exists").
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

	if err := uc.departmentRepo.Create(ctx, department); err != nil {
		return 0, err
	}

	return department.ID, nil
}

// List returns a list of departments with pagination and filtering.
func (uc *DepartmentUseCase) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*Department, int, error) {
	departments, count, err := uc.departmentRepo.List(ctx, filters, limit, offset, sortClause)
	if err != nil {
		return nil, 0, err
	}

	records := make([]*Department, 0, len(departments))
	for _, dept := range departments {
		userCount, err := uc.userRepo.Count(ctx, map[string]any{"department_id": dept.ID, "enabled": true}, nil)
		if err != nil {
			return nil, 0, err
		}

		records = append(records, &Department{
			Department: dept,
			UserCount:  userCount,
		})
	}

	return records, count, nil
}

// GetByID retrieves a department by ID.
func (uc *DepartmentUseCase) GetByID(ctx context.Context, id uint) (*Department, error) {
	department, err := uc.departmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	userCount, err := uc.userRepo.Count(ctx, map[string]any{"department_id": id, "enabled": true}, nil)
	if err != nil {
		return nil, err
	}

	return &Department{
		Department: department,
		UserCount:  userCount,
	}, nil
}

// Update updates an existing department.
func (uc *DepartmentUseCase) Update(ctx context.Context, params UpdateDepartmentParams) error {
	// Get the existing department
	department, err := uc.departmentRepo.FindByID(ctx, params.ID)
	if err != nil {
		return err
	}

	// Check if department already exists
	exists, err := uc.departmentRepo.IsExists(ctx, map[string]any{"name": params.Name}, map[string]any{"id": params.ID})
	if err != nil {
		return err
	}

	if exists {
		return errs.NewBusiness("department name already exists").
			WithReason(errs.ReasonDuplicate)
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

	if err := uc.departmentRepo.Update(ctx, department); err != nil {
		return err
	}
	return nil
}

// Delete removes a department.
func (uc *DepartmentUseCase) Delete(ctx context.Context, id uint) error {
	_, err := uc.departmentRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if department has users
	userExists, err := uc.userRepo.IsExists(ctx, map[string]any{"department_id": id, "enabled": true}, nil)
	if err != nil {
		return err
	}

	if userExists {
		return errs.NewBusiness("department has users").
			WithReason(errs.ReasonInvalidState)
	}

	if err := uc.departmentRepo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}

// GetDepartmentOptions retrieves a list of departments for options.
func (uc *DepartmentUseCase) GetDepartmentOptions(ctx context.Context) ([]*model.Department, error) {
	departments, _, err := uc.departmentRepo.List(ctx, map[string]any{}, -1, -1, "")
	if err != nil {
		return nil, err
	}

	return departments, nil
}
