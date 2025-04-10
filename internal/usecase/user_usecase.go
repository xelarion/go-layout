// Package usecase contains business logic.
package usecase

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/model/gen"
	"github.com/xelarion/go-layout/internal/util"
	"github.com/xelarion/go-layout/pkg/errs"
)

// CreateUserParams contains all parameters needed to create a user
type CreateUserParams struct {
	Username string
	Email    string
	Password string
	Role     string
}

// UpdateUserParams contains all parameters needed to update a user
type UpdateUserParams struct {
	ID       uint
	Username string
	Email    string
	Password string
	Role     string
	Enabled  bool
	// Fields to track which values are explicitly set
	UsernameSet bool
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
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint) error
	SetRSAPrivateKey(ctx context.Context, cacheKey string, privateKey []byte) error
	GetRSAPrivateKey(ctx context.Context, cacheKey string) ([]byte, error)
	DeleteRSAPrivateKey(ctx context.Context, cacheKey string) error
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
	_, err := uc.repo.FindByUsername(ctx, params.Username)
	if err != nil {
		if !errs.IsReason(err, errs.ReasonNotFound) {
			return nil, err
		}
	} else {
		return nil, errs.NewBusiness("username already exists").
			WithReason(errs.ReasonDuplicate)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errs.WrapInternal(err, "failed to hash password")
	}

	// Create user
	user := &model.User{
		User: gen.User{
			Username: params.Username,
			Email:    params.Email,
			Password: string(hashedPassword),
			Role:     params.Role,
		},
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
	if params.UsernameSet {
		user.Username = params.Username
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
func (uc *UserUseCase) Login(ctx context.Context, username, password string) (*model.User, error) {
	user, err := uc.repo.FindByUsername(ctx, username)
	if err != nil {
		if errs.IsReason(err, errs.ReasonNotFound) {
			return nil, errs.NewBusiness("Incorrect username or password").WithReason(errs.ReasonUnauthorized)
		}
		return nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errs.NewBusiness("Incorrect username or password").WithReason(errs.ReasonUnauthorized)
	}

	// Check if user is enabled
	if !user.Enabled {
		return nil, errs.NewBusiness("User account is disabled").WithReason(errs.ReasonUnauthorized)
	}

	return user, nil
}

func (uc *UserUseCase) GetRSAPublicKey(ctx context.Context) (string, string, error) {
	privateKey, publicKey, err := util.GenRSAKey(2048)
	if err != nil {
		return "", "", err
	}

	rdsKey := util.RandSeq(25)
	err = uc.repo.SetRSAPrivateKey(ctx, rdsKey, privateKey)
	if err != nil {
		return "", "", err
	}

	return string(publicKey), rdsKey, nil
}

func (uc *UserUseCase) GetRSAPrivateKey(ctx context.Context, cacheKey string) ([]byte, error) {
	return uc.repo.GetRSAPrivateKey(ctx, cacheKey)
}

func (uc *UserUseCase) DeleteRSAPrivateKey(ctx context.Context, cacheKey string) error {
	return uc.repo.DeleteRSAPrivateKey(ctx, cacheKey)
}
