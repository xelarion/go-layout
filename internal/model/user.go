// Package model contains domain models.
package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user model.
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	Email     string         `gorm:"size:100;uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"size:100;not null" json:"-"` // Password is not exposed in JSON responses
	Role      string         `gorm:"size:20;default:'user'" json:"role"`
	CreatedAt time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // Soft delete
}

// TableName returns the table name for the User model.
func (User) TableName() string {
	return "users"
}
