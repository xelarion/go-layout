package model

import (
	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model/gen"
)

// Role represents a Role
type Role struct {
	gen.Role
}

// IsSuperAdmin returns true if the user has admin role.
func (u *Role) IsSuperAdmin() bool {
	return u.Slug == enum.RoleSuperAdmin
}
