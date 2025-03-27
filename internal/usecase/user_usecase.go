// Package usecase contains business logic.
package usecase

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/xelarion/go-layout/internal/model"
)

var (
	// ErrUserNotFound indicates that the requested user was not found.
	ErrUserNotFound = errors.New("user not found")
	// ErrEmailAlreadyExists indicates that the email is already registered.
	ErrEmailAlreadyExists = errors.New("email already exists")
	// ErrInvalidCredentials indicates that the provided credentials are invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// UserRepository defines methods for user data access.
type UserRepository interface {
	FindByID(ctx context.Context, id uint) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]*model.User, error)
}

// UserUseCase implements business logic for user operations.
type UserUseCase struct {
	repo   UserRepository
	logger *zap.Logger
}

// NewUserUseCase creates a new instance of UserUseCase.
func NewUserUseCase(repo UserRepository, logger *zap.Logger) *UserUseCase {
	return &UserUseCase{
		repo:   repo,
		logger: logger.Named("user_usecase"),
	}
}

// GetByID retrieves a user by ID.
func (uc *UserUseCase) GetByID(ctx context.Context, id uint) (*model.User, error) {
	user, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to find user by ID", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}
	return user, nil
}

// Register creates a new user.
func (uc *UserUseCase) Register(ctx context.Context, name, email, password string) (*model.User, error) {
	// Check if user already exists
	existingUser, err := uc.repo.FindByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		uc.logger.Error("Failed to hash password", zap.Error(err))
		return nil, err
	}

	// Create user
	user := &model.User{
		Name:      name,
		Email:     email,
		Password:  string(hashedPassword),
		Role:      "user", // Default role
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		uc.logger.Error("Failed to create user", zap.String("email", email), zap.Error(err))
		return nil, err
	}

	return user, nil
}

// Login authenticates a user.
func (uc *UserUseCase) Login(ctx context.Context, email, password string) (*model.User, error) {
	user, err := uc.repo.FindByEmail(ctx, email)
	if err != nil {
		uc.logger.Warn("User not found during login attempt", zap.String("email", email), zap.Error(err))
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		uc.logger.Warn("Invalid password during login attempt", zap.String("email", email), zap.Error(err))
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// Update updates an existing user.
func (uc *UserUseCase) Update(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now()
	if err := uc.repo.Update(ctx, user); err != nil {
		uc.logger.Error("Failed to update user", zap.Uint("id", user.ID), zap.Error(err))
		return err
	}
	return nil
}

// Delete removes a user.
func (uc *UserUseCase) Delete(ctx context.Context, id uint) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		uc.logger.Error("Failed to delete user", zap.Uint("id", id), zap.Error(err))
		return err
	}
	return nil
}

// List returns a list of users with pagination.
func (uc *UserUseCase) List(ctx context.Context, limit, offset int) ([]*model.User, error) {
	users, err := uc.repo.List(ctx, limit, offset)
	if err != nil {
		uc.logger.Error("Failed to list users", zap.Int("limit", limit), zap.Int("offset", offset), zap.Error(err))
		return nil, err
	}
	return users, nil
}
