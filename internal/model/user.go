// Package model contains domain models.
package model

import (
	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model/gen"
)

// User represents a user model.
type User struct {
	gen.User
}

// IsAdmin returns true if the user has admin role.
func (u *User) IsAdmin() bool {
	return u.Role == enum.RoleAdmin
}
