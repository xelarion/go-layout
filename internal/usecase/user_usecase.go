// Package usecase contains business logic.
package usecase

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/pkg/errs"
)

// CreateUserParams contains all parameters needed to create a user
type CreateUserParams struct {
	Name     string
	Email    string
	Password string
	Role     string
}

// UpdateUserParams contains all parameters needed to update a user
type UpdateUserParams struct {
	ID       uint
	Name     string
	Email    string
	Password string
	Role     string
	Enabled  bool
	// Fields to track which values are explicitly set
	NameSet     bool
	EmailSet    bool
	PasswordSet bool
	RoleSet     bool
	EnabledSet  bool
}

// UserRepository defines methods for user data access.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.User, int, error)
	FindByID(ctx context.Context, id uint) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint) error
}

// UserUseCase implements business logic for user operations.
type UserUseCase struct {
	repo UserRepository
}

// NewUserUseCase creates a new instance of UserUseCase.
func NewUserUseCase(repo UserRepository) *UserUseCase {
	return &UserUseCase{
		repo: repo,
	}
}

// Create creates a new user.
func (uc *UserUseCase) Create(ctx context.Context, params CreateUserParams) (*model.User, error) {
	// Check if user already exists
	_, err := uc.repo.FindByEmail(ctx, params.Email)
	if err != nil {
		if !errs.IsReason(err, errs.ReasonNotFound) {
			return nil, err
		}
	} else {
		return nil, errs.NewBusiness("email already exists").
			WithReason(errs.ReasonDuplicate)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errs.WrapInternal(err, "failed to hash password")
	}

	// Create user
	user := &model.User{
		Name:     params.Name,
		Email:    params.Email,
		Password: string(hashedPassword),
		Role:     params.Role,
		Enabled:  true,
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// List returns a list of users with pagination and filtering.
func (uc *UserUseCase) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.User, int, error) {
	users, count, err := uc.repo.List(ctx, filters, limit, offset, sortClause)
	if err != nil {
		return nil, 0, err
	}
	return users, count, nil
}

// GetByID retrieves a user by ID.
func (uc *UserUseCase) GetByID(ctx context.Context, id uint) (*model.User, error) {
	user, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Update updates an existing user.
func (uc *UserUseCase) Update(ctx context.Context, params UpdateUserParams) error {
	// Get the existing user
	user, err := uc.repo.FindByID(ctx, params.ID)
	if err != nil {
		return err
	}

	// Update fields that are explicitly set
	if params.NameSet {
		user.Name = params.Name
	}

	if params.EmailSet {
		user.Email = params.Email
	}

	if params.PasswordSet && params.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
		if err != nil {
			return errs.WrapInternal(err, "failed to hash password")
		}
		user.Password = string(hashedPassword)
	}

	if params.RoleSet {
		user.Role = params.Role
	}

	if params.EnabledSet {
		user.Enabled = params.Enabled
	}

	if err := uc.repo.Update(ctx, user); err != nil {
		return err
	}
	return nil
}

// Delete removes a user.
func (uc *UserUseCase) Delete(ctx context.Context, id uint) error {
	_, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}

// Login authenticates a user.
func (uc *UserUseCase) Login(ctx context.Context, email, password string) (*model.User, error) {
	user, err := uc.repo.FindByEmail(ctx, email)
	if err != nil {
		if errs.IsReason(err, errs.ReasonNotFound) {
			return nil, errs.NewBusiness("invalid credentials").WithReason(errs.ReasonUnauthorized)
		}
		return nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errs.NewBusiness("invalid credentials").WithReason(errs.ReasonUnauthorized)
	}

	// Check if user is enabled
	if !user.Enabled {
		return nil, errs.NewBusiness("account is disabled").WithReason(errs.ReasonUnauthorized)
	}

	return user, nil
}
