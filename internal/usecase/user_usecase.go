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

type User struct {
	*model.User
	RoleName       string
	RoleSlug       string
	DepartmentName string
}

// CreateUserParams contains all parameters needed to create a user
type CreateUserParams struct {
	Username     string
	Password     string
	FullName     string
	PhoneNumber  string
	Email        string
	Enabled      bool
	DepartmentID uint
	RoleID       uint
}

// UpdateUserParams contains all parameters needed to update a user
type UpdateUserParams struct {
	ID           uint
	Username     string
	Password     string
	FullName     string
	PhoneNumber  string
	Email        string
	Enabled      bool
	DepartmentID uint
	RoleID       uint

	// Fields to track which values are explicitly set
	UsernameSet     bool
	PasswordSet     bool
	FullNameSet     bool
	PhoneNumberSet  bool
	EmailSet        bool
	EnabledSet      bool
	DepartmentIDSet bool
	RoleIDSet       bool
}

// UserRepository defines methods for user data access.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.User, int, error)
	IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error)
	FindByID(ctx context.Context, id uint) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint) error
	SetRSAPrivateKey(ctx context.Context, cacheKey string, privateKey []byte) error
	GetRSAPrivateKey(ctx context.Context, cacheKey string) ([]byte, error)
	DeleteRSAPrivateKey(ctx context.Context, cacheKey string) error
}

// UserUseCase implements business logic for user operations.
type UserUseCase struct {
	userRepo       UserRepository
	roleRepo       RoleRepository
	departmentRepo DepartmentRepository
}

// NewUserUseCase creates a new instance of UserUseCase.
func NewUserUseCase(repo UserRepository, roleRepo RoleRepository, departmentRepo DepartmentRepository) *UserUseCase {
	return &UserUseCase{
		userRepo:       repo,
		roleRepo:       roleRepo,
		departmentRepo: departmentRepo,
	}
}

// Create creates a new user.
func (uc *UserUseCase) Create(ctx context.Context, params CreateUserParams) (uint, error) {
	// Check if user already exists
	exists, err := uc.userRepo.IsExists(ctx, map[string]any{"username": params.Username}, nil)
	if err != nil {
		return 0, err
	}

	if exists {
		return 0, errs.NewBusiness("username already exists").
			WithReason(errs.ReasonDuplicate)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, errs.WrapInternal(err, "failed to hash password")
	}

	// Create user
	user := &model.User{
		User: gen.User{
			Username:     params.Username,
			Password:     string(hashedPassword),
			FullName:     params.FullName,
			PhoneNumber:  params.PhoneNumber,
			Email:        params.Email,
			Enabled:      params.Enabled,
			DepartmentID: params.DepartmentID,
			RoleID:       params.RoleID,
		},
	}

	err = uc.userRepo.Create(ctx, user)
	if err != nil {
		return 0, err
	}

	return user.ID, nil
}

// List returns a list of users with pagination and filtering.
func (uc *UserUseCase) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*User, int, error) {
	users, count, err := uc.userRepo.List(ctx, filters, limit, offset, sortClause)
	if err != nil {
		return nil, 0, err
	}

	var usersWithRoles []*User
	for _, user := range users {
		userWithRoles := &User{
			User: user,
		}

		role, err := uc.roleRepo.FindByID(ctx, user.RoleID)
		if err != nil {
			return nil, 0, err
		}
		userWithRoles.RoleName = role.Name
		userWithRoles.RoleSlug = role.Slug

		department, err := uc.departmentRepo.FindByID(ctx, user.DepartmentID)
		if err != nil {
			if !errs.IsReason(err, errs.ReasonNotFound) {
				return nil, 0, err
			}
		} else {
			userWithRoles.DepartmentName = department.Name
		}

		usersWithRoles = append(usersWithRoles, userWithRoles)
	}

	return usersWithRoles, count, nil
}

// GetByID retrieves a user by ID.
func (uc *UserUseCase) GetByID(ctx context.Context, id uint) (*User, error) {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	role, err := uc.roleRepo.FindByID(ctx, user.RoleID)
	if err != nil {
		return nil, err
	}

	departmentName := ""
	department, err := uc.departmentRepo.FindByID(ctx, user.DepartmentID)
	if err != nil {
		if !errs.IsReason(err, errs.ReasonNotFound) {
			return nil, err
		}
	} else {
		departmentName = department.Name
	}

	return &User{
		User:           user,
		RoleName:       role.Name,
		RoleSlug:       role.Slug,
		DepartmentName: departmentName,
	}, nil
}

// Update updates an existing user.
func (uc *UserUseCase) Update(ctx context.Context, params UpdateUserParams) error {
	// Get the existing user
	user, err := uc.userRepo.FindByID(ctx, params.ID)
	if err != nil {
		return err
	}

	// Check if username already exists
	exists, err := uc.userRepo.IsExists(ctx, map[string]any{"username": params.Username}, map[string]any{"id": params.ID})
	if err != nil {
		return err
	}

	if exists {
		return errs.NewBusiness("username already exists").
			WithReason(errs.ReasonDuplicate)
	}

	// Update fields that are explicitly set
	if params.UsernameSet {
		user.Username = params.Username
	}

	if params.PasswordSet && params.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
		if err != nil {
			return errs.WrapInternal(err, "failed to hash password")
		}
		user.Password = string(hashedPassword)
	}

	if params.FullNameSet {
		user.FullName = params.FullName
	}

	if params.PhoneNumberSet {
		user.PhoneNumber = params.PhoneNumber
	}

	if params.EmailSet {
		user.Email = params.Email
	}

	if params.EnabledSet {
		user.Enabled = params.Enabled
	}

	if params.DepartmentIDSet {
		user.DepartmentID = params.DepartmentID
	}

	if params.RoleIDSet {
		user.RoleID = params.RoleID
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return err
	}
	return nil
}

// Delete removes a user.
func (uc *UserUseCase) Delete(ctx context.Context, id uint) error {
	_, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.userRepo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}

// Login authenticates a user.
func (uc *UserUseCase) Login(ctx context.Context, username, password string) (*User, error) {
	user, err := uc.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if errs.IsReason(err, errs.ReasonNotFound) {
			return nil, errs.NewBusiness("incorrect username or password").WithReason(errs.ReasonUnauthorized)
		}
		return nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errs.NewBusiness("incorrect username or password").WithReason(errs.ReasonUnauthorized)
	}

	// Check if user is enabled
	if !user.Enabled {
		return nil, errs.NewBusiness("user account is disabled").WithReason(errs.ReasonUnauthorized)
	}

	role, err := uc.roleRepo.FindByID(ctx, user.RoleID)
	if err != nil {
		return nil, err
	}

	return &User{
		User:     user,
		RoleName: role.Name,
		RoleSlug: role.Slug,
	}, nil
}

func (uc *UserUseCase) GetRSAPublicKey(ctx context.Context) (string, string, error) {
	privateKey, publicKey, err := util.GenRSAKey(2048)
	if err != nil {
		return "", "", err
	}

	rdsKey := util.RandSeq(25)
	err = uc.userRepo.SetRSAPrivateKey(ctx, rdsKey, privateKey)
	if err != nil {
		return "", "", err
	}

	return string(publicKey), rdsKey, nil
}

func (uc *UserUseCase) GetRSAPrivateKey(ctx context.Context, cacheKey string) ([]byte, error) {
	return uc.userRepo.GetRSAPrivateKey(ctx, cacheKey)
}

func (uc *UserUseCase) DeleteRSAPrivateKey(ctx context.Context, cacheKey string) error {
	return uc.userRepo.DeleteRSAPrivateKey(ctx, cacheKey)
}
