// Package repository provides data access implementations.
package repository

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/pkg/errs"
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

// Create adds a new user to the database.
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return errs.WrapInternal(err, "failed to create user")
	}
	return nil
}

// List retrieves users with pagination and filtering.
func (r *UserRepository) List(ctx context.Context, filters map[string]any, limit, offset int) ([]*model.User, int, error) {
	query := r.db.WithContext(ctx).Model(&model.User{})

	for field, value := range filters {
		if value != nil {
			query = query.Where(field+" = ?", value)
		}
	}

	var total int64
	if err := query.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to count users")
	}

	var users []*model.User
	if err := query.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to list users")
	}

	return users, int(total), nil
}

// FindByID retrieves a user by ID.
func (r *UserRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("user not found").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, errs.WrapInternal(err, "failed to find user by ID")
	}
	return &user, nil
}

// FindByEmail retrieves a user by email.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("user not found").
				WithReason(errs.ReasonNotFound).
				WithMeta("email", email)
		}
		return nil, errs.WrapInternal(err, "failed to find user by email")
	}
	return &user, nil
}

// Update updates a user.
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "failed to update user")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("user not found").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", user.ID)
	}

	return nil
}

// Delete removes a user by ID.
func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.User{}, id)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "failed to delete user")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("user not found").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", id)
	}

	return nil
}
