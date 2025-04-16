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
	DepartmentID uint
	RoleID       uint

	// Fields to track which values are explicitly set
	UsernameSet     bool
	PasswordSet     bool
	FullNameSet     bool
	PhoneNumberSet  bool
	EmailSet        bool
	DepartmentIDSet bool
	RoleIDSet       bool
}

// UserRepository defines methods for user data access.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.User, int, error)
	IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error)
	Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error)
	FindByID(ctx context.Context, id uint) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, user *model.User, params map[string]any) error
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

	// Check is department enabled
	department, err := uc.departmentRepo.FindByID(ctx, params.DepartmentID)
	if err != nil {
		return 0, err
	}

	if !department.Enabled {
		return 0, errs.NewBusiness("department is disabled").
			WithReason(errs.ReasonInvalidState)
	}

	// Check is role enabled
	role, err := uc.roleRepo.FindByID(ctx, params.RoleID)
	if err != nil {
		return 0, err
	}

	if !role.Enabled {
		return 0, errs.NewBusiness("role is disabled").
			WithReason(errs.ReasonInvalidState)
	}

	// can't create super admin user
	if role.IsSuperAdmin() {
		return 0, errs.NewBusiness("can't create super admin user").
			WithReason(errs.ReasonInvalidState)
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

	records := make([]*User, 0, len(users))
	for _, user := range users {
		record := &User{
			User: user,
		}

		role, err := uc.roleRepo.FindByID(ctx, user.RoleID)
		if err != nil {
			return nil, 0, err
		}

		record.RoleName = role.Name
		record.RoleSlug = role.Slug

		department, err := uc.departmentRepo.FindByID(ctx, user.DepartmentID)
		if err != nil {
			if !errs.IsReason(err, errs.ReasonNotFound) {
				return nil, 0, err
			}
		} else {
			record.DepartmentName = department.Name
		}

		records = append(records, record)
	}

	return records, count, nil
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

	updates := map[string]any{}

	// Update fields that are explicitly set
	if params.UsernameSet {
		// Check if username already exists
		exists, err := uc.userRepo.IsExists(ctx, map[string]any{"username": params.Username}, map[string]any{"id": params.ID})
		if err != nil {
			return err
		}

		if exists {
			return errs.NewBusiness("username already exists").
				WithReason(errs.ReasonDuplicate)
		}

		updates["username"] = params.Username
	}

	if params.PasswordSet && params.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
		if err != nil {
			return errs.WrapInternal(err, "failed to hash password")
		}
		updates["password"] = string(hashedPassword)
	}

	if params.FullNameSet {
		updates["full_name"] = params.FullName
	}

	if params.PhoneNumberSet {
		updates["phone_number"] = params.PhoneNumber
	}

	if params.EmailSet {
		updates["email"] = params.Email
	}

	if params.DepartmentIDSet {
		// Check is department enabled
		department, err := uc.departmentRepo.FindByID(ctx, params.DepartmentID)
		if err != nil {
			return err
		}

		if !department.Enabled {
			return errs.NewBusiness("department is disabled").
				WithReason(errs.ReasonInvalidState)
		}

		updates["department_id"] = params.DepartmentID
	}

	if params.RoleIDSet {
		// Check is role enabled
		role, err := uc.roleRepo.FindByID(ctx, params.RoleID)
		if err != nil {
			return err
		}

		if !role.Enabled {
			return errs.NewBusiness("role is disabled").
				WithReason(errs.ReasonInvalidState)
		}

		if user.RoleID != params.RoleID {
			// can't set user's role to super admin
			if role.IsSuperAdmin() {
				return errs.NewBusiness("can't set super admin role").
					WithReason(errs.ReasonInvalidState)
			}

			userOldRole, err := uc.roleRepo.FindByID(ctx, user.RoleID)
			if err != nil {
				return err
			}

			// can't update super admin user's role
			if userOldRole.IsSuperAdmin() {
				return errs.NewBusiness("can't update super admin role").
					WithReason(errs.ReasonInvalidState)
			}
		}

		updates["role_id"] = params.RoleID
	}

	if err := uc.userRepo.Update(ctx, user, updates); err != nil {
		return err
	}

	return nil
}

// UpdateEnabled updates user enabled.
func (uc *UserUseCase) UpdateEnabled(ctx context.Context, id uint, enabled bool) error {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	role, err := uc.roleRepo.FindByID(ctx, user.RoleID)
	if err != nil {
		return err
	}

	// can't update super admin enabled
	if role.IsSuperAdmin() {
		return errs.NewBusiness("can't update super admin enabled").
			WithReason(errs.ReasonInvalidState)
	}

	if err := uc.userRepo.Update(ctx, user, map[string]any{"enabled": enabled}); err != nil {
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
		return nil, errs.NewBusiness("user account is disabled").WithReason(errs.ReasonUserDisabled)
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
