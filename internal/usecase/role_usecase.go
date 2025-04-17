// Package usecase contains business logic.
package usecase

import (
	"context"

	"github.com/lib/pq"

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

	// Fields to track which values are explicitly set
	NameSet        bool
	DescriptionSet bool
}

// RoleRepository defines methods for role data access.
type RoleRepository interface {
	Create(ctx context.Context, role *model.Role) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.Role, int, error)
	IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error)
	Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error)
	FindByID(ctx context.Context, id uint) (*model.Role, error)
	Update(ctx context.Context, role *model.Role, params map[string]any) error
	Delete(ctx context.Context, id uint) error
}

// RoleUseCase implements business logic for role operations.
type RoleUseCase struct {
	tx       Transaction
	roleRepo RoleRepository
	userRepo UserRepository
}

// NewRoleUseCase creates a new instance of RoleUseCase.
func NewRoleUseCase(tx Transaction, repo RoleRepository, userRepo UserRepository) *RoleUseCase {
	return &RoleUseCase{
		tx:       tx,
		roleRepo: repo,
		userRepo: userRepo,
	}
}

// Create creates a new role.
func (uc *RoleUseCase) Create(ctx context.Context, params CreateRoleParams) (uint, error) {
	// Check if role already exists
	exists, err := uc.roleRepo.IsExists(ctx, map[string]any{"name": params.Name}, nil)
	if err != nil {
		return 0, err
	}

	if exists {
		return 0, errs.NewBusiness("role name already exists").
			WithReason(errs.ReasonDuplicate)
	}

	var roleID uint
	err = uc.tx.Transaction(ctx, func(ctx context.Context) error {
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
			return err
		}

		roleID = role.ID
		return nil
	})

	if err != nil {
		return 0, err
	}

	return roleID, nil
}

// List returns a list of roles with pagination and filtering.
func (uc *RoleUseCase) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*Role, int, error) {
	roles, count, err := uc.roleRepo.List(ctx, filters, limit, offset, sortClause)
	if err != nil {
		return nil, 0, err
	}

	records := make([]*Role, 0, len(roles))
	for _, role := range roles {
		userCount, err := uc.userRepo.Count(ctx, map[string]any{"role_id": role.ID, "enabled": true}, nil)
		if err != nil {
			return nil, 0, err
		}

		records = append(records, &Role{
			Role:      role,
			UserCount: userCount,
		})
	}

	return records, count, nil
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

	updates := map[string]any{}

	// Update fields that are explicitly set
	if params.NameSet {
		// Check if role already exists
		exists, err := uc.roleRepo.IsExists(ctx, map[string]any{"name": params.Name}, map[string]any{"id": params.ID})
		if err != nil {
			return err
		}

		if exists {
			return errs.NewBusiness("role name already exists").
				WithReason(errs.ReasonDuplicate)
		}

		updates["name"] = params.Name
	}

	if params.DescriptionSet {
		updates["description"] = params.Description
	}

	return uc.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := uc.roleRepo.Update(ctx, role, updates); err != nil {
			return err
		}
		return nil
	})
}

// UpdateEnabled updates role enabled.
func (uc *RoleUseCase) UpdateEnabled(ctx context.Context, id uint, enabled bool) error {
	// Get the existing role
	role, err := uc.roleRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// super admin cannot be updated
	if role.IsSuperAdmin() {
		return errs.NewBusiness("super admin cannot be updated").
			WithReason(errs.ReasonInvalidState)
	}

	return uc.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := uc.roleRepo.Update(ctx, role, map[string]any{"enabled": enabled}); err != nil {
			return err
		}
		return nil
	})
}

// Delete removes a role.
func (uc *RoleUseCase) Delete(ctx context.Context, id uint) error {
	// Check if role exists
	_, err := uc.roleRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if role has users
	userExists, err := uc.userRepo.IsExists(ctx, map[string]any{"role_id": id, "enabled": true}, nil)
	if err != nil {
		return err
	}

	if userExists {
		return errs.NewBusiness("role has users").
			WithReason(errs.ReasonInvalidState)
	}

	return uc.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := uc.roleRepo.Delete(ctx, id); err != nil {
			return err
		}
		return nil
	})
}

// GetRoleOptions retrieves a list of roles for options.
func (uc *RoleUseCase) GetRoleOptions(ctx context.Context) ([]*model.Role, error) {
	roles, _, err := uc.roleRepo.List(ctx, map[string]any{}, -1, -1, "")
	if err != nil {
		return nil, err
	}

	return roles, nil
}

// UpdatePermissions updates role permissions
func (uc *RoleUseCase) UpdatePermissions(ctx context.Context, roleID uint, permissions []string) error {
	// Get the existing role
	role, err := uc.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return err
	}

	// super admin cannot be updated
	if role.IsSuperAdmin() {
		return errs.NewBusiness("super admin permissions cannot be updated").
			WithReason(errs.ReasonInvalidState)
	}

	return uc.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := uc.roleRepo.Update(ctx, role, map[string]any{"permissions": pq.StringArray(permissions)}); err != nil {
			return err
		}
		return nil
	})
}
