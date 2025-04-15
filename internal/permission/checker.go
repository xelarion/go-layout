package permission

import (
	"github.com/xelarion/go-layout/internal/enum"
)

// RolePermissionChecker provides methods to check permissions for a role
type RolePermissionChecker struct {
	Slug        string
	Permissions []string
}

// NewRolePermissionChecker creates a new permission checker for a role
func NewRolePermissionChecker(slug string, permissions []string) *RolePermissionChecker {
	return &RolePermissionChecker{
		Slug:        slug,
		Permissions: permissions,
	}
}

// IsSuperAdmin returns true if the role is a super admin
func (c *RolePermissionChecker) IsSuperAdmin() bool {
	return c.Slug == enum.RoleSuperAdmin
}

// HasPermission checks if the role has a specific permission
func (c *RolePermissionChecker) HasPermission(permission string) bool {
	// Super admin has all permissions
	if c.IsSuperAdmin() {
		return true
	}

	// Check if the role has the specific permission
	for _, p := range c.Permissions {
		// Check for wildcard permission or exact match
		if p == All || p == permission {
			return true
		}
	}

	return false
}

// HasAnyPermission checks if the role has any of the given permissions
func (c *RolePermissionChecker) HasAnyPermission(permissions ...string) bool {
	// Super admin has all permissions
	if c.IsSuperAdmin() {
		return true
	}

	// Check if the role has any of the permissions
	for _, permission := range permissions {
		if c.HasPermission(permission) {
			return true
		}
	}

	return false
}
