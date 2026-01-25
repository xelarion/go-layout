// Package usecase contains business logic.
package usecase

import (
	"context"
	"fmt"
	"time"

	excelexporter "github.com/xelarion/excel-exporter"
	"golang.org/x/crypto/bcrypt"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/model/gen"
	"github.com/xelarion/go-layout/internal/util"
	"github.com/xelarion/go-layout/pkg/errs"
)

type User struct {
	*model.User
	RoleName         string
	RoleSlug         string
	OrganizationName string
	PositionName     string
}

// CreateUserParams contains all parameters needed to create a user
type CreateUserParams struct {
	CreateRecordCommonParams
	FullName    string
	EmployeeNo  string
	Username    string
	Gender      enum.UserGender
	PositionID  uint
	PhoneNumber string
	Email       string
	RoleID      uint
	Enabled     bool
	Remark      string
}

// UpdateUserParams contains all parameters needed to update a user
type UpdateUserParams struct {
	UpdateRecordCommonParams
	ID          uint
	FullName    string
	EmployeeNo  string
	Username    string
	Password    string
	Gender      enum.UserGender
	PositionID  uint
	PhoneNumber string
	Email       string
	RoleID      uint
	Enabled     bool
	Remark      string

	// Fields to track which values are explicitly set
	FullNameSet    bool
	EmployeeNoSet  bool
	UsernameSet    bool
	PasswordSet    bool
	GenderSet      bool
	PositionIDSet  bool
	PhoneNumberSet bool
	EmailSet       bool
	RoleIDSet      bool
	EnabledSet     bool
	RemarkSet      bool
}

// BatchDeleteUsersResult represents the result of a single user deletion.
type BatchDeleteUsersResult = BatchItemCommonResult

// UserRepository defines methods for user data access.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.User, int, error)
	IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error)
	Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error)
	CountByOrganizationID(ctx context.Context, organizationID uint) (int64, error)
	FindByID(ctx context.Context, id uint) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, user *model.User, params map[string]any) error
	Delete(ctx context.Context, id uint) error
}

// UserUseCase implements business logic for user operations.
type UserUseCase struct {
	tx               Transaction
	userRepo         UserRepository
	cacheRepo        CacheRepository
	roleRepo         RoleRepository
	organizationRepo OrganizationRepository
	positionRepo     PositionRepository
}

// NewUserUseCase creates a new instance of UserUseCase.
func NewUserUseCase(
	tx Transaction,
	userRepo UserRepository,
	cacheRepo CacheRepository,
	roleRepo RoleRepository,
	organizationRepo OrganizationRepository,
	positionRepo PositionRepository,
) *UserUseCase {
	return &UserUseCase{
		tx:               tx,
		userRepo:         userRepo,
		cacheRepo:        cacheRepo,
		roleRepo:         roleRepo,
		organizationRepo: organizationRepo,
		positionRepo:     positionRepo,
	}
}

// Create creates a new user.
func (uc *UserUseCase) Create(ctx context.Context, params CreateUserParams) (uint, error) {
	// Check if employee no already exists
	if exists, err := uc.userRepo.IsExists(ctx, map[string]any{"employee_no": params.EmployeeNo}, nil); err != nil {
		return 0, err
	} else if exists {
		return 0, errs.NewBusiness("工号已存在").
			WithReason(errs.ReasonDuplicate)
	}

	// Check if username already exists
	if exists, err := uc.userRepo.IsExists(ctx, map[string]any{"username": params.Username}, nil); err != nil {
		return 0, err
	} else if exists {
		return 0, errs.NewBusiness("用户名已存在").
			WithReason(errs.ReasonDuplicate)
	}

	// Check if phone number already exists
	if exists, err := uc.userRepo.IsExists(ctx, map[string]any{"phone_number": params.PhoneNumber}, nil); err != nil {
		return 0, err
	} else if exists {
		return 0, errs.NewBusiness("手机号已存在").
			WithReason(errs.ReasonDuplicate)
	}

	// Check if email already exists
	if exists, err := uc.userRepo.IsExists(ctx, map[string]any{"email": params.Email}, nil); err != nil {
		return 0, err
	} else if exists {
		return 0, errs.NewBusiness("邮箱已存在").
			WithReason(errs.ReasonDuplicate)
	}

	// Check is role enabled
	role, err := uc.roleRepo.FindByID(ctx, params.RoleID)
	if err != nil {
		return 0, err
	}

	if !role.Enabled {
		return 0, errs.NewBusiness("角色已禁用").
			WithReason(errs.ReasonInvalidState)
	}

	// can't create super admin user
	if role.IsSuperAdmin() {
		return 0, errs.NewBusiness("不能创建超级管理员").
			WithReason(errs.ReasonInvalidState)
	}

	// Hash password
	password := "123456" // TODO random initial password for sms or email
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, errs.WrapInternal(err, "failed to hash password")
	}

	user := &model.User{
		User: gen.User{
			CreatedBy:     params.CreatedBy,
			UpdatedBy:     params.CreatedBy,
			CreatedRoleID: params.CreatedRoleID, // 设置创建时的角色
			FullName:      params.FullName,
			EmployeeNo:    params.EmployeeNo,
			Username:      params.Username,
			Password:      string(hashedPassword),
			PhoneNumber:   params.PhoneNumber,
			Email:         params.Email,
			Enabled:       params.Enabled,
			Remark:        params.Remark,
			PositionID:    params.PositionID,
			RoleID:        params.RoleID,
		},
		Gender: params.Gender,
	}
	err = uc.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := uc.userRepo.Create(ctx, user); err != nil {
			return err
		}
		return nil
	})
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

		position, err := uc.positionRepo.FindByID(ctx, user.PositionID)
		if err != nil {
			if !errs.IsReason(err, errs.ReasonNotFound) {
				return nil, 0, err
			}
			record.PositionName = ""
		} else {
			record.PositionName = position.Name
		}

		if position != nil && position.OrganizationID > 0 {
			organization, err := uc.organizationRepo.FindByID(ctx, position.OrganizationID)
			if err != nil {
				if !errs.IsReason(err, errs.ReasonNotFound) {
					return nil, 0, err
				}
				record.OrganizationName = ""
			} else {
				record.OrganizationName = organization.Name
			}
		}

		records = append(records, record)
	}

	return records, count, nil
}

// Count returns the total count of users matching the filters.
func (uc *UserUseCase) Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error) {
	return uc.userRepo.Count(ctx, filters, notFilters)
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

	positionName := ""
	position, err := uc.positionRepo.FindByID(ctx, user.PositionID)
	if err != nil {
		if !errs.IsReason(err, errs.ReasonNotFound) {
			return nil, err
		}
	} else {
		positionName = position.Name
	}

	organizationName := ""
	if position != nil && position.OrganizationID > 0 {
		organization, err := uc.organizationRepo.FindByID(ctx, position.OrganizationID)
		if err != nil {
			if !errs.IsReason(err, errs.ReasonNotFound) {
				return nil, err
			}
		} else {
			organizationName = organization.Name
		}
	}

	return &User{
		User:             user,
		RoleName:         role.Name,
		RoleSlug:         role.Slug,
		OrganizationName: organizationName,
		PositionName:     positionName,
	}, nil
}

// Update updates an existing user.
func (uc *UserUseCase) Update(ctx context.Context, params UpdateUserParams) error {
	// Get the existing user
	user, err := uc.userRepo.FindByID(ctx, params.ID)
	if err != nil {
		return err
	}

	userOldRole, err := uc.roleRepo.FindByID(ctx, user.RoleID)
	if err != nil {
		return err
	}

	updates := map[string]any{}

	if params.FullNameSet {
		updates["full_name"] = params.FullName
	}

	if params.EmployeeNoSet {
		// Check if employee no already exists
		if exists, err := uc.userRepo.IsExists(ctx, map[string]any{"employee_no": params.EmployeeNo}, map[string]any{"id": params.ID}); err != nil {
			return err
		} else if exists {
			return errs.NewBusiness("工号已存在").
				WithReason(errs.ReasonDuplicate)
		}

		updates["employee_no"] = params.EmployeeNo
	}

	// Update fields that are explicitly set
	if params.UsernameSet {
		// Check if username already exists
		if exists, err := uc.userRepo.IsExists(ctx, map[string]any{"username": params.Username}, map[string]any{"id": params.ID}); err != nil {
			return err
		} else if exists {
			return errs.NewBusiness("用户名已存在").
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

	if params.GenderSet {
		updates["gender"] = params.Gender
	}

	if params.PositionIDSet {
		updates["position_id"] = params.PositionID
	}

	if params.PhoneNumberSet {
		// Check if phone number already exists
		if exists, err := uc.userRepo.IsExists(ctx, map[string]any{"phone_number": params.PhoneNumber}, map[string]any{"id": params.ID}); err != nil {
			return err
		} else if exists {
			return errs.NewBusiness("手机号已存在").
				WithReason(errs.ReasonDuplicate)
		}

		updates["phone_number"] = params.PhoneNumber
	}

	if params.EmailSet {
		// Check if email already exists
		if exists, err := uc.userRepo.IsExists(ctx, map[string]any{"email": params.Email}, map[string]any{"id": params.ID}); err != nil {
			return err
		} else if exists {
			return errs.NewBusiness("邮箱已存在").
				WithReason(errs.ReasonDuplicate)
		}

		updates["email"] = params.Email
	}

	if params.RoleIDSet {
		// Check is role enabled
		role, err := uc.roleRepo.FindByID(ctx, params.RoleID)
		if err != nil {
			return err
		}

		if !role.Enabled {
			return errs.NewBusiness("角色已被禁用").
				WithReason(errs.ReasonInvalidState)
		}

		if user.RoleID != params.RoleID {
			// can't set user's role to super admin
			if role.IsSuperAdmin() {
				return errs.NewBusiness("不能将用户角色设置为超级管理员").
					WithReason(errs.ReasonInvalidState)
			}

			// can't update super admin user's role
			if userOldRole.IsSuperAdmin() {
				return errs.NewBusiness("不能修改超级管理员角色").
					WithReason(errs.ReasonInvalidState)
			}
		}

		updates["role_id"] = params.RoleID
	}

	if params.EnabledSet {
		// can't update super admin enabled
		if userOldRole.IsSuperAdmin() && params.Enabled != userOldRole.Enabled {
			return errs.NewBusiness("不能修改超级管理员状态").
				WithReason(errs.ReasonInvalidState)
		}
		updates["enabled"] = params.Enabled
	}

	if params.RemarkSet {
		updates["remark"] = params.Remark
	}

	if len(updates) > 0 {
		updates["updated_by"] = params.UpdatedBy
		err = uc.tx.Transaction(ctx, func(ctx context.Context) error {
			if err := uc.userRepo.Update(ctx, user, updates); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// Delete removes a user.
func (uc *UserUseCase) Delete(ctx context.Context, id uint) error {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	role, err := uc.roleRepo.FindByID(ctx, user.RoleID)
	if err != nil {
		return err
	}

	// can't delete super admin
	if role.IsSuperAdmin() {
		return errs.NewBusiness("不能删除超级管理员").
			WithReason(errs.ReasonInvalidState)
	}

	// 检查是否有业务关联
	if exists, err := uc.organizationRepo.IsExists(ctx, map[string]any{"owner_id": id}, nil); err != nil {
		return err
	} else if exists {
		return errs.NewBusiness("该用户已被设置组织负责人，无法删除").
			WithReason(errs.ReasonInvalidState)
	}

	return uc.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := uc.userRepo.Delete(ctx, id); err != nil {
			return err
		}
		return nil
	})
}

// BatchDelete removes multiple users and returns detailed results.
func (uc *UserUseCase) BatchDelete(ctx context.Context, ids []uint) []BatchDeleteUsersResult {
	results := make([]BatchDeleteUsersResult, 0, len(ids))

	for _, id := range ids {
		result := BatchDeleteUsersResult{
			ID:      id,
			Success: false,
		}

		// Get user info first for better error messages
		user, err := uc.userRepo.FindByID(ctx, id)
		if err != nil {
			result.Identifier = fmt.Sprintf("ID %d 未知记录", id)
			result.Error = err
			results = append(results, result)
			continue
		}

		result.Identifier = user.FullName

		// Attempt to delete
		err = uc.Delete(ctx, id)
		if err != nil {
			result.Error = err
		} else {
			result.Success = true
		}

		results = append(results, result)
	}

	return results
}

// GetUserOptions 获取用户作为选项数据
func (uc *UserUseCase) GetUserOptions(ctx context.Context, filters map[string]any, sortClause string) ([]*User, error) {
	users, _, err := uc.userRepo.List(ctx, filters, -1, -1, sortClause)
	if err != nil {
		return nil, err
	}

	records := make([]*User, 0, len(users))
	for _, user := range users {
		record := &User{
			User: user,
		}
		records = append(records, record)
	}

	return records, nil
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
func (uc *UserUseCase) CheckCurrentPassword(ctx context.Context, id uint, currentPassword string) error {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errs.IsReason(err, errs.ReasonNotFound) {
			return errs.NewBusiness("未找到用户")
		}
		return err
	}
	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return errs.NewBusiness("旧密码错误")
	}
	return nil
}
func (uc *UserUseCase) GetRSAPublicKey(ctx context.Context) (string, string, error) {
	privateKey, publicKey, err := util.GenRSAKey(2048)
	if err != nil {
		return "", "", err
	}

	rdsKey := util.RandSeq(25)
	// 使用 CacheRepository 存储临时 RSA 私钥，10 分钟过期
	err = uc.cacheRepo.Set(ctx, rdsKey, privateKey, 10*time.Minute)
	if err != nil {
		return "", "", err
	}

	return string(publicKey), rdsKey, nil
}

func (uc *UserUseCase) GetRSAPrivateKey(ctx context.Context, cacheKey string) ([]byte, error) {
	val, err := uc.cacheRepo.Get(ctx, cacheKey)
	if err != nil {
		if errs.IsReason(err, errs.ReasonNotFound) {
			return nil, errs.NewBusiness("页面已过期，请刷新页面")
		}
		return nil, err
	}
	return val, nil
}

func (uc *UserUseCase) DeleteRSAPrivateKey(ctx context.Context, cacheKey string) error {
	return uc.cacheRepo.Delete(ctx, cacheKey)
}

// ExportUsers 导出用户数据到Excel文件
func (uc *UserUseCase) ExportUsers(ctx context.Context, filters map[string]any, filePath string) error {
	// 使用 List 获取准确的总数（适配 ids、key 等复杂查询条件）
	_, totalCount, err := uc.List(ctx, filters, 0, 0, "")
	if err != nil {
		return errs.WrapInternal(err, "failed to count users")
	}

	batchSize := 500
	e := excelexporter.New(filePath, true)
	sheets := []excelexporter.SheetData{
		{
			Name: "用户列表",
			RowFunc: excelexporter.UseRowChan(func(dataCh chan excelexporter.Row) error {
				// 表头
				dataCh <- excelexporter.NewRow("序号", "用户姓名", "工号", "用户名", "性别", "组织", "岗位", "手机号", "邮箱", "角色", "状态")

				// 分批查询并发送数据
				seq := 1
				for offset := 0; offset < totalCount; offset += batchSize {
					users, _, err := uc.List(ctx, filters, batchSize, offset, "id DESC")
					if err != nil {
						return errs.WrapInternal(err, "failed to query users")
					}

					for _, user := range users {
						enabledText := "禁用"
						if user.Enabled {
							enabledText = "启用"
						}
						dataCh <- excelexporter.NewRow(
							seq,
							user.FullName,
							user.EmployeeNo,
							user.Username,
							enum.UserGenderText[user.Gender],
							user.OrganizationName,
							user.PositionName,
							user.PhoneNumber,
							user.Email,
							user.RoleName,
							enabledText,
						)
						seq++
					}
				}
				return nil
			}),
		},
	}

	return e.Export(sheets)
}
