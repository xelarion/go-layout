package model

import (
	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model/gen"
	"github.com/xelarion/go-layout/internal/permission"
)

// Role represents a Role
type Role struct {
	gen.Role
}

// IsSuperAdmin returns true if the user has admin role.
func (u *Role) IsSuperAdmin() bool {
	return u.Slug == enum.RoleSuperAdmin
}

// HasPermission checks if the role has specific permission
func (u *Role) HasPermission(perm string) bool {
	checker := permission.NewRolePermissionChecker(u.Slug, u.Permissions)
	return checker.HasPermission(perm)
}

// HasAnyPermission checks if the role has any of the given permissions
func (u *Role) HasAnyPermission(perms ...string) bool {
	checker := permission.NewRolePermissionChecker(u.Slug, u.Permissions)
	return checker.HasAnyPermission(perms...)
}
