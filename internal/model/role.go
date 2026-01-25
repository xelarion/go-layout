package model

import (
	"gorm.io/datatypes"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model/gen"
	"github.com/xelarion/go-layout/internal/permission"
)

// Role represents a Role
type Role struct {
	gen.Role

	DataScopes datatypes.JSONSlice[permission.RoleDataScope] `gorm:"column:data_scopes;type:jsonb;not null;default:[];comment:Role data scopes" json:"data_scopes"` // Role data scopes
}

// IsSuperAdmin returns true if the user has admin role.
func (r *Role) IsSuperAdmin() bool {
	return r.Slug == enum.RoleSuperAdmin
}

// GetResourceType 实现 Resource 接口
func (r *Role) GetResourceType() string {
	return "role"
}

// GetResourceID 实现 Resource 接口
func (r *Role) GetResourceID() uint {
	return r.ID
}

// HasPermission checks if the role has specific permission
func (r *Role) HasPermission(perm string) bool {
	checker := permission.NewRoleChecker(r.Slug, r.Permissions)
	return checker.Has(perm)
}

// HasAnyPermission checks if the role has any of the given permissions
func (r *Role) HasAnyPermission(perms ...string) bool {
	checker := permission.NewRoleChecker(r.Slug, r.Permissions)
	return checker.HasAny(perms...)
}

// GetDataScopes 实现 RoleWithDataScopes 接口
func (r *Role) GetDataScopes() []permission.RoleDataScope {
	return r.DataScopes
}

// GetCreatedBy 实现 DataScopeRecord 接口
func (r *Role) GetCreatedBy() uint {
	return r.CreatedBy
}

// GetCreatedRoleID 实现 DataScopeRecord 接口
func (r *Role) GetCreatedRoleID() uint {
	return r.CreatedRoleID
}

// GetAssignedTo 实现 DataScopeRecord 接口
func (r *Role) GetAssignedTo() *uint {
	return nil // roles 表没有 assigned_to 字段
}

// GetDataScopeFieldMap 实现 DataScopeRecord 接口
// 返回 roles 表支持的数据权限字段映射
func (r *Role) GetDataScopeFieldMap() map[string]string {
	return map[string]string{
		"created_by":      "created_by",      // 创建人字段
		"created_role_id": "created_role_id", // 创建角色字段
		// roles 表不支持 assigned_to，所以不在此定义
	}
}
