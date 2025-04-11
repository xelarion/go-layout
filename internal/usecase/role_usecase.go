// Package usecase contains business logic.
package usecase

import (
	"context"

	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/model/gen"
	"github.com/xelarion/go-layout/pkg/errs"
)

// Role contains role data
type Role struct {
	*model.Role
	UserCount int64
}

// CreateRoleParams contains all parameters needed to create a role
type CreateRoleParams struct {
	Name        string
	Description string
	Enabled     bool
}

// UpdateRoleParams contains all parameters needed to update a role
type UpdateRoleParams struct {
	ID          uint
	Name        string
	Description string
	Enabled     bool

	// Fields to track which values are explicitly set
	NameSet        bool
	DescriptionSet bool
	EnabledSet     bool
}

// RoleRepository defines methods for role data access.
type RoleRepository interface {
	Create(ctx context.Context, role *model.Role) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.Role, int, error)
	FindByID(ctx context.Context, id uint) (*model.Role, error)
	FindByName(ctx context.Context, name string) (*model.Role, error)
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id uint) error
	CountUsersByRoleID(ctx context.Context, roleID uint) (int64, error)
}

// RoleUseCase implements business logic for role operations.
type RoleUseCase struct {
	roleRepo RoleRepository
}

// NewRoleUseCase creates a new instance of RoleUseCase.
func NewRoleUseCase(repo RoleRepository) *RoleUseCase {
	return &RoleUseCase{
		roleRepo: repo,
	}
}

// Create creates a new role.
func (uc *RoleUseCase) Create(ctx context.Context, params CreateRoleParams) (uint, error) {
	// Check if role already exists
	_, err := uc.roleRepo.FindByName(ctx, params.Name)
	if err != nil {
		if !errs.IsReason(err, errs.ReasonNotFound) {
			return 0, err
		}
	} else {
		return 0, errs.NewBusiness("role name already exists").
			WithReason(errs.ReasonDuplicate)
	}

	// Create role
	role := &model.Role{
		Role: gen.Role{
			Name:        params.Name,
			Slug:        "",
			Description: params.Description,
			Enabled:     params.Enabled,
			Permissions: []string{},
		},
	}

	if err := uc.roleRepo.Create(ctx, role); err != nil {
		return 0, err
	}

	return role.ID, nil
}

// List returns a list of roles with pagination and filtering.
func (uc *RoleUseCase) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*Role, int, error) {
	roles, count, err := uc.roleRepo.List(ctx, filters, limit, offset, sortClause)
	if err != nil {
		return nil, 0, err
	}

	// Get user count for each role
	rolesWithUserCount := make([]*Role, 0, len(roles))
	for _, dept := range roles {
		userCount, err := uc.roleRepo.CountUsersByRoleID(ctx, dept.ID)
		if err != nil {
			return nil, 0, err
		}
		rolesWithUserCount = append(rolesWithUserCount, &Role{
			Role:      dept,
			UserCount: userCount,
		})
	}

	return rolesWithUserCount, count, nil
}

// GetByID retrieves a role by ID.
func (uc *RoleUseCase) GetByID(ctx context.Context, id uint) (*Role, error) {
	role, err := uc.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &Role{
		Role: role,
	}, nil
}

// Update updates an existing role.
func (uc *RoleUseCase) Update(ctx context.Context, params UpdateRoleParams) error {
	// Get the existing role
	role, err := uc.roleRepo.FindByID(ctx, params.ID)
	if err != nil {
		return err
	}

	// Update fields that are explicitly set
	if params.NameSet {
		role.Name = params.Name
	}

	if params.DescriptionSet {
		role.Description = params.Description
	}

	if params.EnabledSet {
		role.Enabled = params.Enabled
	}

	if err := uc.roleRepo.Update(ctx, role); err != nil {
		return err
	}
	return nil
}

// Delete removes a role.
func (uc *RoleUseCase) Delete(ctx context.Context, id uint) error {
	_, err := uc.roleRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.roleRepo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}

// GetRoleOptions retrieves a list of roles for options.
func (uc *RoleUseCase) GetRoleOptions(ctx context.Context) ([]*model.Role, error) {
	roles, _, err := uc.roleRepo.List(ctx, map[string]any{}, -1, -1, "")
	if err != nil {
		return nil, err
	}

	return roles, nil
}
