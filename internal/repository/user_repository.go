// Package repository provides data access implementations.
package repository

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/pkg/errs"
)

// UserRepository is an implementation of the user repository.
type UserRepository struct {
	db  *gorm.DB
	rds *redis.Client
}

// NewUserRepository creates a new instance of user repository.
func NewUserRepository(db *gorm.DB, rds *redis.Client) *UserRepository {
	return &UserRepository{
		db:  db,
		rds: rds,
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
func (r *UserRepository) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.User, int, error) {
	query := r.db.WithContext(ctx).Model(&model.User{})

	for field, value := range filters {
		if value == nil {
			continue
		}

		switch field {
		case "username":
			if str, ok := value.(string); ok {
				query = query.Where(field+" LIKE ?", "%"+str+"%")
			}
		default:
			query = query.Where(field+" = ?", value)
		}
	}

	var total int64
	if err := query.Model(&model.User{}).Count(&total).Error; err != nil {
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

// FindByUsername retrieves a user by username.
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
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

// SetRSAPrivateKey sets the RSA private key in the cache.
func (r *UserRepository) SetRSAPrivateKey(ctx context.Context, cacheKey string, privateKey []byte) error {
	err := r.rds.Set(ctx, cacheKey, privateKey, 0).Err()
	if err != nil {
		return errs.WrapInternal(err, "failed to set RSA private key in cache")
	}
	return nil
}

func (r *UserRepository) GetRSAPrivateKey(ctx context.Context, cacheKey string) ([]byte, error) {
	key, err := r.rds.Get(ctx, cacheKey).Bytes()
	if err != nil {
		return nil, errs.WrapInternal(err, "failed to get RSA private key from cache")
	}
	return key, nil
}

// DeleteRSAPrivateKey removes the RSA private key from the cache.
func (r *UserRepository) DeleteRSAPrivateKey(ctx context.Context, cacheKey string) error {
	err := r.rds.Del(ctx, cacheKey).Err()
	if err != nil {
		return errs.WrapInternal(err, "failed to delete RSA private key from cache")
	}
	return nil
}
