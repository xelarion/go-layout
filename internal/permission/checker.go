package permission

import (
	"github.com/xelarion/go-layout/internal/enum"
)

// RoleChecker provides methods to check permissions for a role
type RoleChecker struct {
	Slug        string
	Permissions []string
}

// NewRoleChecker creates a new permission checker for a role
func NewRoleChecker(slug string, permissions []string) *RoleChecker {
	return &RoleChecker{
		Slug:        slug,
		Permissions: permissions,
	}
}

// IsSuperAdmin returns true if the role is a super admin
func (c *RoleChecker) IsSuperAdmin() bool {
	return c.Slug == enum.RoleSuperAdmin
}

// Has checks if the role has a specific permission
func (c *RoleChecker) Has(permission string) bool {
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

// HasAny checks if the role has any of the given permissions
func (c *RoleChecker) HasAny(permissions ...string) bool {
	// Super admin has all permissions
	if c.IsSuperAdmin() {
		return true
	}

	// Check if the role has any of the permissions
	for _, permission := range permissions {
		if c.Has(permission) {
			return true
		}
	}

	return false
}
