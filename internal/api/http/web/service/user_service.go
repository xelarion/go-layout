package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	appctx "github.com/xelarion/go-layout/internal/context"
	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/internal/util"
	"github.com/xelarion/go-layout/pkg/errs"
)

// UserService handles user-related services.
type UserService struct {
	*BaseService  // 嵌入 BaseService 提供数据权限检查方法
	userUseCase   *usecase.UserUseCase
	exportUseCase *usecase.ExportUseCase
	logger        *zap.Logger
}

// NewUserService creates a new UserService.
func NewUserService(
	baseService *BaseService,
	userUseCase *usecase.UserUseCase,
	exportUseCase *usecase.ExportUseCase,
	logger *zap.Logger,
) *UserService {
	return &UserService{
		BaseService:   baseService,
		userUseCase:   userUseCase,
		exportUseCase: exportUseCase,
		logger:        logger,
	}
}

// CreateUser registers a new user.
func (s *UserService) CreateUser(ctx context.Context, req *types.CreateUserReq) (*types.CreateUserResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	gender, err := enum.GetUserGender(req.Gender)
	if err != nil {
		return nil, err
	}

	params := usecase.CreateUserParams{
		CreateRecordCommonParams: usecase.CreateRecordCommonParams{
			CreatedBy:     current.UserID,
			CreatedRoleID: current.RoleID,
			AssignedTo:    current.UserID,
		},
		FullName:    req.FullName,
		EmployeeNo:  req.EmployeeNo,
		Username:    req.Username,
		Gender:      gender,
		PositionID:  req.PositionID,
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
		RoleID:      req.RoleID,
		Enabled:     req.Enabled,
		Remark:      req.Remark,
	}

	id, err := s.userUseCase.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	return &types.CreateUserResp{
		ID: id,
	}, nil
}

// ListUsers lists users with pagination.
func (s *UserService) ListUsers(ctx context.Context, req *types.ListUsersReq) (*types.ListUsersResp, error) {
	filters := map[string]any{}
	if req.Key != "" {
		filters["key"] = req.Key
	}

	if req.RoleID != 0 {
		filters["role_id"] = req.RoleID
	}

	if req.OrganizationID != 0 {
		filters["organization_id"] = req.OrganizationID
	}

	if req.Enabled != nil {
		filters["enabled"] = *req.Enabled
	}

	// 应用数据权限过滤（传入空的 User 实例用于获取字段映射）
	if err := s.ApplyDataScope(ctx, filters, &model.User{}); err != nil {
		return nil, err
	}

	users, count, err := s.userUseCase.List(ctx, filters, req.GetLimit(), req.GetOffset(), req.GetSortClause())
	if err != nil {
		return nil, err
	}

	respResults := make([]types.ListUsersRespResult, 0, len(users))
	for _, user := range users {
		u := types.ListUsersRespResult{
			ID:               user.ID,
			FullName:         user.FullName,
			EmployeeNo:       user.EmployeeNo,
			Username:         user.Username,
			Gender:           enum.UserGenderText[user.Gender],
			OrganizationName: user.OrganizationName,
			PositionName:     user.PositionName,
			PhoneNumber:      user.PhoneNumber,
			Email:            user.Email,
			RoleName:         user.RoleName,
			RoleSlug:         user.RoleSlug,
			Enabled:          user.Enabled,
		}
		respResults = append(respResults, u)
	}

	return &types.ListUsersResp{
		Results:    respResults,
		Pagination: types.NewPageResp(count, req.GetPage(), req.GetPageSize()),
	}, nil
}

// GetUser retrieves a user by ID.
func (s *UserService) GetUser(ctx context.Context, req *types.GetUserReq) (*types.GetUserResp, error) {
	user, err := s.userUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 检查数据权限
	if err := s.CheckRecordAccess(ctx, user.User); err != nil {
		return nil, err
	}

	return &types.GetUserResp{
		ID:               user.ID,
		CreatedAt:        util.FormatDateTime(user.CreatedAt),
		FullName:         user.FullName,
		EmployeeNo:       user.EmployeeNo,
		Username:         user.Username,
		Gender:           enum.UserGenderText[user.Gender],
		OrganizationName: user.OrganizationName,
		PositionName:     user.PositionName,
		PhoneNumber:      user.PhoneNumber,
		Email:            user.Email,
		RoleName:         user.RoleName,
		RoleSlug:         user.RoleSlug,
		Enabled:          user.Enabled,
	}, nil
}

// GetUserFormData provides data needed for user forms (update).
func (s *UserService) GetUserFormData(ctx context.Context, req *types.GetUserFormDataReq) (*types.GetUserFormDataResp, error) {
	user, err := s.userUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 检查数据权限
	if err := s.CheckRecordAccess(ctx, user.User); err != nil {
		return nil, err
	}

	return &types.GetUserFormDataResp{
		ID:          user.ID,
		FullName:    user.FullName,
		EmployeeNo:  user.EmployeeNo,
		Username:    user.Username,
		Gender:      enum.UserGenderSlug[user.Gender],
		PositionID:  user.PositionID,
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
		RoleID:      user.RoleID,
		RoleSlug:    user.RoleSlug,
		Enabled:     user.Enabled,
		Remark:      user.Remark,
	}, nil
}

// UpdateUser updates a user.
func (s *UserService) UpdateUser(ctx context.Context, req *types.UpdateUserReq) (*types.UpdateUserResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	// 先获取用户检查数据权限
	user, err := s.userUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	if err := s.CheckRecordAccess(ctx, user.User); err != nil {
		return nil, err
	}

	// Create update params
	params := usecase.UpdateUserParams{
		UpdateRecordCommonParams: usecase.UpdateRecordCommonParams{
			UpdatedBy: current.UserID,
		},
		ID: req.ID,
	}

	params.FullName = req.FullName
	params.FullNameSet = true

	params.EmployeeNo = req.EmployeeNo
	params.EmployeeNoSet = true

	params.Username = req.Username
	params.UsernameSet = true

	params.Gender, err = enum.GetUserGender(req.Gender)
	if err != nil {
		return nil, err
	}
	params.GenderSet = true

	params.PositionID = req.PositionID
	params.PositionIDSet = true

	params.PhoneNumber = req.PhoneNumber
	params.PhoneNumberSet = true

	params.Email = req.Email
	params.EmailSet = true

	params.RoleID = req.RoleID
	params.RoleIDSet = true

	params.Enabled = req.Enabled
	params.EnabledSet = true

	params.Remark = req.Remark
	params.RemarkSet = true

	if err := s.userUseCase.Update(ctx, params); err != nil {
		return nil, err
	}

	return &types.UpdateUserResp{}, nil
}

// UpdateUserEnabled updates a user's enabled status.
func (s *UserService) UpdateUserEnabled(ctx context.Context, req *types.UpdateUserEnabledReq) (*types.UpdateUserResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	// 先获取用户检查数据权限
	user, err := s.userUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	if err := s.CheckRecordAccess(ctx, user.User); err != nil {
		return nil, err
	}

	params := usecase.UpdateUserParams{
		UpdateRecordCommonParams: usecase.UpdateRecordCommonParams{
			UpdatedBy: current.UserID,
		},
		ID: req.ID,
	}
	params.Enabled = *req.Enabled
	params.EnabledSet = true
	if err := s.userUseCase.Update(ctx, params); err != nil {
		return nil, err
	}

	return &types.UpdateUserResp{}, nil
}

// UpdateUserPassword updates a user's password.
func (s *UserService) UpdateUserPassword(ctx context.Context, req *types.UpdateUserPasswordReq) (*types.UpdateUserPasswordResp, error) {
	// 先获取用户检查数据权限
	user, err := s.userUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	if err := s.CheckRecordAccess(ctx, user.User); err != nil {
		return nil, err
	}

	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	// Create update params
	params := usecase.UpdateUserParams{
		UpdateRecordCommonParams: usecase.UpdateRecordCommonParams{
			UpdatedBy: current.UserID,
		},
		ID: req.ID,
	}

	params.Password = req.Password
	params.PasswordSet = true

	if err := s.userUseCase.Update(ctx, params); err != nil {
		return nil, err
	}

	return &types.UpdateUserPasswordResp{}, nil
}

// DeleteUser handles user deletion.
func (s *UserService) DeleteUser(ctx context.Context, req *types.DeleteUserReq) (*types.DeleteUserResp, error) {
	user, err := s.userUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 检查数据权限
	if err := s.CheckRecordAccess(ctx, user.User); err != nil {
		return nil, err
	}

	if err := s.userUseCase.Delete(ctx, req.ID); err != nil {
		return nil, err
	}

	return &types.DeleteUserResp{}, nil
}

// BatchDeleteUsers handles batch user deletion.
func (s *UserService) BatchDeleteUsers(ctx context.Context, req *types.BatchDeleteUsersReq) (*types.BatchDeleteUsersResp, error) {
	// 先检查每个角色的数据权限
	accessibleIDs := make([]uint, 0, len(req.IDs))
	failedItems := make([]string, 0)

	for _, id := range req.IDs {
		user, err := s.userUseCase.GetByID(ctx, id)
		if err != nil {
			// 用户不存在
			failedItems = append(failedItems, fmt.Sprintf("用户ID %d: 用户不存在或已被删除", id))
			continue
		}

		// 检查数据权限
		if err := s.CheckRecordAccess(ctx, user.User); err != nil {
			failedItems = append(failedItems, user.FullName+": 无权限删除此用户")
			continue
		}

		accessibleIDs = append(accessibleIDs, id)
	}

	// 批量删除有权限的用户
	results := s.userUseCase.BatchDelete(ctx, accessibleIDs)

	// 统计结果
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			// 格式化错误信息：用户名: 错误原因
			errMsg := result.Identifier + ": " + errs.GetMessage(result.Error)
			failedItems = append(failedItems, errMsg)
		}
	}

	return &types.BatchDeleteUsersResp{
		SuccessCount: successCount,
		FailedCount:  len(failedItems),
		FailedItems:  failedItems,
	}, nil
}

func (s *UserService) GetUserOptions(ctx context.Context, req *types.GetUserOptionsReq) (*types.GetUserOptionsResp, error) {
	filters := map[string]any{}
	if req.Enabled != nil {
		filters["enabled"] = *req.Enabled
	}

	users, err := s.userUseCase.GetUserOptions(ctx, filters, req.GetSortClause())
	if err != nil {
		return nil, err
	}

	results := make([]types.GetUserOptionsRespResult, 0, len(users))
	for _, user := range users {
		results = append(results, types.GetUserOptionsRespResult{
			GetOptionsRespResult: types.GetOptionsRespResult{
				Label: user.FullName,
				Value: user.ID,
			},
			Enabled: user.Enabled,
		})
	}

	return &types.GetUserOptionsResp{
		Results: results,
	}, nil
}

// ExportUsers 导出用户列表（支持同步/异步）
func (s *UserService) ExportUsers(ctx context.Context, req *types.ExportUsersReq) (*types.ExportResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	// 构建筛选条件
	filters := map[string]any{}
	if len(req.IDs) > 0 {
		filters["ids"] = req.IDs
	}

	if req.Key != "" {
		filters["key"] = req.Key
	}

	if req.RoleID != 0 {
		filters["role_id"] = req.RoleID
	}

	if req.OrganizationID != 0 {
		filters["organization_id"] = req.OrganizationID
	}

	if req.Enabled != nil {
		filters["enabled"] = *req.Enabled
	}

	// 应用数据权限过滤
	if err := s.ApplyDataScope(ctx, filters, &model.User{}); err != nil {
		return nil, err
	}

	// 查询总数(使用 List 获取准确的总数（适配 ids、key 等复杂查询条件）)
	_, totalCount, err := s.userUseCase.List(ctx, filters, 0, 0, "")
	if err != nil {
		return nil, err
	}

	if totalCount == 0 {
		return nil, errs.NewBusiness("没有可导出的用户数据").
			WithReason(errs.ReasonNotFound)
	}

	filename := fmt.Sprintf("用户列表-%s.xlsx", time.Now().Format("20060102150405"))
	result, err := s.exportUseCase.Execute(
		ctx,
		enum.ExportTaskTypeUser,
		totalCount,
		filters,
		current.UserID,
		filename,
		s.userUseCase.ExportUsers,
	)
	if err != nil {
		return nil, err
	}

	return s.NewExportResp(result), nil
}
