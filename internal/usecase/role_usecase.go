// Package usecase contains business logic.
package usecase

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/datatypes"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/model/gen"
	"github.com/xelarion/go-layout/internal/permission"
	"github.com/xelarion/go-layout/pkg/errs"
)

// Role contains role data
type Role struct {
	*model.Role
	UserCount int64
}

// CreateRoleParams contains all parameters needed to create a role
type CreateRoleParams struct {
	CreateRecordCommonParams
	Code         string
	Name         string
	DisplayOrder int32
	Description  string
	Enabled      bool
}

// UpdateRoleParams contains all parameters needed to update a role
type UpdateRoleParams struct {
	UpdateRecordCommonParams
	ID           uint
	Code         string
	Name         string
	DisplayOrder int32
	Description  string
	Enabled      bool

	// Fields to track which values are explicitly set
	CodeSet         bool
	NameSet         bool
	DisplayOrderSet bool
	DescriptionSet  bool
	EnabledSet      bool
}

// UpdatePermissionsParams contains all parameters needed to update a role permissions
type UpdatePermissionsParams struct {
	UpdateRecordCommonParams
	ID          uint
	Permissions []string
	DataScopes  []UpdatePermissionsDataScope
}

// UpdatePermissionsDataScope contains all parameters needed to update a role permissions data scope
type UpdatePermissionsDataScope struct {
	Permissions []string `json:"permissions"`
	Type        string   `json:"type"`
}

// BatchDeleteRolesResult represents the result of a single role deletion.
type BatchDeleteRolesResult = BatchItemCommonResult

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
	if exists, err := uc.roleRepo.IsExists(ctx, map[string]any{"code": params.Code}, nil); err != nil {
		return 0, err
	} else if exists {
		return 0, errs.NewBusiness("角色编号已存在").
			WithReason(errs.ReasonDuplicate)
	}

	// Check if role already exists
	if exists, err := uc.roleRepo.IsExists(ctx, map[string]any{"name": params.Name}, nil); err != nil {
		return 0, err
	} else if exists {
		return 0, errs.NewBusiness("角色名称已存在").
			WithReason(errs.ReasonDuplicate)
	}

	role := &model.Role{
		Role: gen.Role{
			CreatedBy:     params.CreatedBy,
			UpdatedBy:     params.CreatedBy,
			CreatedRoleID: params.CreatedRoleID, // 设置创建时的角色
			Name:          params.Name,
			Slug:          "",
			Code:          params.Code,
			DisplayOrder:  params.DisplayOrder,
			Description:   params.Description,
			Enabled:       params.Enabled,
			Permissions:   []string{},
		},
		DataScopes: datatypes.NewJSONSlice([]permission.RoleDataScope{}),
	}
	err := uc.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := uc.roleRepo.Create(ctx, role); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
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

	records := make([]*Role, 0, len(roles))
	for _, role := range roles {
		userCount, err := uc.userRepo.Count(ctx, map[string]any{"role_id": role.ID}, nil)
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
	if params.CodeSet {
		// Check if role already exists
		exists, err := uc.roleRepo.IsExists(ctx, map[string]any{"code": params.Code}, map[string]any{"id": params.ID})
		if err != nil {
			return err
		}

		if exists {
			return errs.NewBusiness("角色编号已存在").
				WithReason(errs.ReasonDuplicate)
		}

		updates["code"] = params.Code
	}

	if params.NameSet {
		// Check if role already exists
		exists, err := uc.roleRepo.IsExists(ctx, map[string]any{"name": params.Name}, map[string]any{"id": params.ID})
		if err != nil {
			return err
		}

		if exists {
			return errs.NewBusiness("角色名称已存在").
				WithReason(errs.ReasonDuplicate)
		}

		updates["name"] = params.Name
	}

	if params.DisplayOrderSet {
		updates["display_order"] = params.DisplayOrder
	}

	if params.DescriptionSet {
		updates["description"] = params.Description
	}

	if params.EnabledSet {
		// 超级管理员禁止更新状态
		if role.IsSuperAdmin() && params.Enabled != role.Enabled {
			return errs.NewBusiness("超级管理员角色禁止修改状态").
				WithReason(errs.ReasonForbidden)
		}
		updates["enabled"] = params.Enabled
	}

	if len(updates) > 0 {
		updates["updated_by"] = params.UpdatedBy
		err = uc.tx.Transaction(ctx, func(ctx context.Context) error {
			if err := uc.roleRepo.Update(ctx, role, updates); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// Delete removes a role.
func (uc *RoleUseCase) Delete(ctx context.Context, id uint) error {
	// Check if role exists
	role, err := uc.roleRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// super admin cannot be updated
	if role.IsSuperAdmin() {
		return errs.NewBusiness("超级管理员角色禁止删除").
			WithReason(errs.ReasonForbidden)
	}

	// Check if role has users
	userExists, err := uc.userRepo.IsExists(ctx, map[string]any{"role_id": id}, nil)
	if err != nil {
		return err
	}

	if userExists {
		return errs.NewBusiness("角色下存在用户，无法删除").
			WithReason(errs.ReasonInvalidState)
	}

	return uc.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := uc.roleRepo.Delete(ctx, id); err != nil {
			return err
		}
		return nil
	})
}

// BatchDelete removes multiple roles and returns detailed results.
func (uc *RoleUseCase) BatchDelete(ctx context.Context, ids []uint) []BatchDeleteRolesResult {
	results := make([]BatchDeleteRolesResult, 0, len(ids))

	for _, id := range ids {
		result := BatchDeleteRolesResult{
			ID:      id,
			Success: false,
		}

		// Get role info first for better error messages
		role, err := uc.roleRepo.FindByID(ctx, id)
		if err != nil {
			result.Identifier = fmt.Sprintf("ID %d 未知记录", id)
			result.Error = err
			results = append(results, result)
			continue
		}

		result.Identifier = role.Name

		// Attempt to delete
		err = uc.Delete(ctx, id)
		if err != nil {
			result.Error = err
		} else {
			result.Success = true
		}

		results = append(results, result)
	}

	return results
}

// GetRoleOptions 获取角色作为选项数据
func (uc *RoleUseCase) GetRoleOptions(ctx context.Context, filters map[string]any, sortClause string) ([]*Role, error) {
	roles, _, err := uc.roleRepo.List(ctx, filters, -1, -1, sortClause)
	if err != nil {
		return nil, err
	}

	records := make([]*Role, 0, len(roles))
	for _, role := range roles {
		record := &Role{
			Role: role,
		}
		records = append(records, record)
	}

	return records, nil
}

// UpdatePermissions updates role permissions
func (uc *RoleUseCase) UpdatePermissions(ctx context.Context, params UpdatePermissionsParams) error {
	// Get the existing role
	role, err := uc.roleRepo.FindByID(ctx, params.ID)
	if err != nil {
		return err
	}

	// super admin cannot be updated
	if role.IsSuperAdmin() {
		return errs.NewBusiness("超级管理员角色禁止更新权限").
			WithReason(errs.ReasonInvalidState)
	}

	var dataScopes datatypes.JSONSlice[permission.RoleDataScope]
	for _, dataScope := range params.DataScopes {
		dataScopes = append(dataScopes, permission.RoleDataScope{
			Permissions: dataScope.Permissions,
			Type:        enum.RoleDataScopeType(dataScope.Type),
		})
	}

	return uc.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := uc.roleRepo.Update(ctx, role, map[string]any{
			"updated_by":  params.UpdatedBy,
			"permissions": pq.StringArray(params.Permissions),
			"data_scopes": dataScopes,
		}); err != nil {
			return err
		}
		return nil
	})
}
