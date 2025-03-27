// Package repository provides data access implementations.
package repository

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/xelarion/go-layout/internal/model"
)

// UserRepository is a PostgreSQL implementation of the user repository.
type UserRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewUserRepository creates a new instance of user repository.
func NewUserRepository(db *gorm.DB, logger *zap.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger.Named("user_repository"),
	}
}

// FindByID retrieves a user by ID.
func (r *UserRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("Failed to find user by ID", zap.Uint("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}
	return &user, nil
}

// FindByEmail retrieves a user by email.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("Failed to find user by email", zap.String("email", email), zap.Error(err))
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return &user, nil
}

// Create creates a new user.
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		r.logger.Error("Failed to create user", zap.String("email", user.Email), zap.Error(err))
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// Update updates an existing user.
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		r.logger.Error("Failed to update user", zap.Uint("id", user.ID), zap.Error(err))
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// Delete deletes a user by ID.
func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.User{}, id).Error; err != nil {
		r.logger.Error("Failed to delete user", zap.Uint("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// List retrieves a list of users with pagination.
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*model.User, error) {
	var users []*model.User
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		r.logger.Error("Failed to list users", zap.Int("limit", limit), zap.Int("offset", offset), zap.Error(err))
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}
