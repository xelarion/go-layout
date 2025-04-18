// Package repository provides data access implementations.
package repository

import (
	"context"
	"database/sql"
	"errors"

	"gorm.io/gorm"

	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/errs"
)

var _ usecase.UserRepository = (*UserRepository)(nil)

// UserRepository is an implementation of the user repository.
type UserRepository struct {
	data *Data
}

// NewUserRepository creates a new instance of user repository.
func NewUserRepository(data *Data) *UserRepository {
	return &UserRepository{
		data: data,
	}
}

// Create adds a new user to the database.
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	if err := r.data.DB(ctx).Create(user).Error; err != nil {
		return errs.WrapInternal(err, "failed to create user")
	}
	return nil
}

// List retrieves users with pagination and filtering.
func (r *UserRepository) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.User, int, error) {
	query := r.data.DB(ctx).Model(&model.User{})

	for field, value := range filters {
		if value == nil {
			continue
		}

		switch field {
		case "key":
			if str, ok := value.(string); ok {
				query = query.Where("users.username LIKE @key OR users.full_name LIKE @key", sql.Named("key", "%"+str+"%"))
			}
		default:
			query = query.Where("users."+field+" = ?", value)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to count users")
	}

	if sortClause != "" {
		query = query.Order(sortClause)
	} else {
		query = query.Order("users.id desc")
	}

	var users []*model.User
	if err := query.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to list users")
	}

	return users, int(total), nil
}

func (r *UserRepository) IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error) {
	return IsExists(ctx, r.data.DB(ctx), &model.User{}, filters, notFilters)
}

func (r *UserRepository) Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error) {
	return Count(ctx, r.data.DB(ctx), &model.User{}, filters, notFilters)
}

// FindByID retrieves a user by ID.
func (r *UserRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := r.data.DB(ctx).First(&user, id).Error
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

// FindByUsername retrieves a user by username.
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.data.DB(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("user not found").
				WithReason(errs.ReasonNotFound).
				WithMeta("username", username)
		}
		return nil, errs.WrapInternal(err, "failed to find user by username")
	}
	return &user, nil
}

// Update updates a user.
func (r *UserRepository) Update(ctx context.Context, user *model.User, params map[string]any) error {
	result := r.data.DB(ctx).Model(user).Updates(params)
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
	result := r.data.DB(ctx).Delete(&model.User{}, id)
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

// SetRSAPrivateKey sets the RSA private key in the cache.
func (r *UserRepository) SetRSAPrivateKey(ctx context.Context, cacheKey string, privateKey []byte) error {
	err := r.data.rds.Set(ctx, cacheKey, privateKey, 0).Err()
	if err != nil {
		return errs.WrapInternal(err, "failed to set RSA private key in cache")
	}
	return nil
}

func (r *UserRepository) GetRSAPrivateKey(ctx context.Context, cacheKey string) ([]byte, error) {
	key, err := r.data.rds.Get(ctx, cacheKey).Bytes()
	if err != nil {
		return nil, errs.WrapInternal(err, "failed to get RSA private key from cache")
	}
	return key, nil
}

// DeleteRSAPrivateKey removes the RSA private key from the cache.
func (r *UserRepository) DeleteRSAPrivateKey(ctx context.Context, cacheKey string) error {
	err := r.data.rds.Del(ctx, cacheKey).Err()
	if err != nil {
		return errs.WrapInternal(err, "failed to delete RSA private key from cache")
	}
	return nil
}
