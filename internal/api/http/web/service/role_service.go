package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	appctx "github.com/xelarion/go-layout/internal/context"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/internal/util"
	"github.com/xelarion/go-layout/pkg/errs"
)

// RoleService handles role-related services.
type RoleService struct {
	*BaseService // 嵌入 BaseService 提供数据权限检查方法
	roleUseCase  *usecase.RoleUseCase
	logger       *zap.Logger
}

// NewRoleService creates a new RoleService.
func NewRoleService(
	baseService *BaseService,
	roleUseCase *usecase.RoleUseCase,
	logger *zap.Logger,
) *RoleService {
	return &RoleService{
		BaseService: baseService,
		roleUseCase: roleUseCase,
		logger:      logger,
	}
}

// CreateRole registers a new role.
func (s *RoleService) CreateRole(ctx context.Context, req *types.CreateRoleReq) (*types.CreateRoleResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	params := usecase.CreateRoleParams{
		CreateRecordCommonParams: usecase.CreateRecordCommonParams{
			CreatedBy:     current.UserID,
			CreatedRoleID: current.RoleID,
			AssignedTo:    0, // roles 表不使用 assigned_to
		},
		Code:         req.Code,
		Name:         req.Name,
		DisplayOrder: req.DisplayOrder,
		Description:  req.Description,
		Enabled:      req.Enabled,
	}

	id, err := s.roleUseCase.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	return &types.CreateRoleResp{
		ID: id,
	}, nil
}

// ListRoles lists roles with pagination.
func (s *RoleService) ListRoles(ctx context.Context, req *types.ListRolesReq) (*types.ListRolesResp, error) {
	filters := map[string]any{}
	if req.Key != "" {
		filters["key"] = req.Key
	}

	if req.Enabled != nil {
		filters["enabled"] = *req.Enabled
	}

	// 应用数据权限过滤（传入空的 Role 实例用于获取字段映射）
	if err := s.ApplyDataScope(ctx, filters, &model.Role{}); err != nil {
		return nil, err
	}

	roles, count, err := s.roleUseCase.List(ctx, filters, req.GetLimit(), req.GetOffset(), req.GetSortClause())
	if err != nil {
		return nil, err
	}

	respResults := make([]types.ListRolesRespResult, 0, len(roles))
	for _, role := range roles {
		u := types.ListRolesRespResult{
			ID:           role.ID,
			Name:         role.Name,
			Slug:         role.Slug,
			Code:         role.Code,
			DisplayOrder: role.DisplayOrder,
			Description:  role.Description,
			Enabled:      role.Enabled,
			UserCount:    role.UserCount,
		}
		respResults = append(respResults, u)
	}

	return &types.ListRolesResp{
		Results:    respResults,
		Pagination: types.NewPageResp(count, req.GetPage(), req.GetPageSize()),
	}, nil
}

// GetRole retrieves a role by ID.
func (s *RoleService) GetRole(ctx context.Context, req *types.GetRoleReq) (*types.GetRoleResp, error) {
	role, err := s.roleUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 检查数据权限
	if err := s.CheckRecordAccess(ctx, role.Role); err != nil {
		return nil, err
	}

	return &types.GetRoleResp{
		ID:           role.ID,
		Code:         role.Code,
		Slug:         role.Slug,
		Name:         role.Name,
		DisplayOrder: role.DisplayOrder,
		Description:  role.Description,
		Enabled:      role.Enabled,
		CreatedAt:    util.FormatDateTime(role.CreatedAt),
		UserCount:    role.UserCount,
	}, nil
}

// GetRoleFormData provides data needed for role forms (update).
func (s *RoleService) GetRoleFormData(ctx context.Context, req *types.GetRoleFormDataReq) (*types.GetRoleFormDataResp, error) {
	role, err := s.roleUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 检查数据权限
	if err := s.CheckRecordAccess(ctx, role.Role); err != nil {
		return nil, err
	}

	return &types.GetRoleFormDataResp{
		ID:           role.ID,
		Code:         role.Code,
		Slug:         role.Slug,
		Name:         role.Name,
		DisplayOrder: role.DisplayOrder,
		Description:  role.Description,
		Enabled:      role.Enabled,
	}, nil
}

// UpdateRole updates a role.
func (s *RoleService) UpdateRole(ctx context.Context, req *types.UpdateRoleReq) (*types.UpdateRoleResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	// 先获取角色，用于权限检查
	role, err := s.roleUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 检查数据权限
	if err := s.CheckRecordAccess(ctx, role.Role); err != nil {
		return nil, err
	}

	// Create update params
	params := usecase.UpdateRoleParams{
		UpdateRecordCommonParams: usecase.UpdateRecordCommonParams{
			UpdatedBy: current.UserID,
		},
		ID: req.ID,
	}

	params.Code = req.Code
	params.CodeSet = true

	params.Name = req.Name
	params.NameSet = true

	params.DisplayOrder = req.DisplayOrder
	params.DisplayOrderSet = true

	params.Description = req.Description
	params.DescriptionSet = true

	params.Enabled = req.Enabled
	params.EnabledSet = true

	if err := s.roleUseCase.Update(ctx, params); err != nil {
		return nil, err
	}

	return &types.UpdateRoleResp{}, nil
}

// UpdateRoleEnabled updates a role's enabled status.
func (s *RoleService) UpdateRoleEnabled(ctx context.Context, req *types.UpdateRoleEnabledReq) (*types.UpdateRoleResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	// 先获取角色，用于权限检查
	role, err := s.roleUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 检查数据权限
	if err := s.CheckRecordAccess(ctx, role.Role); err != nil {
		return nil, err
	}

	// Create update params
	params := usecase.UpdateRoleParams{
		UpdateRecordCommonParams: usecase.UpdateRecordCommonParams{
			UpdatedBy: current.UserID,
		},
		ID: req.ID,
	}
	params.Enabled = *req.Enabled
	params.EnabledSet = true
	if err := s.roleUseCase.Update(ctx, params); err != nil {
		return nil, err
	}

	return &types.UpdateRoleResp{}, nil
}

// DeleteRole handles role deletion.
func (s *RoleService) DeleteRole(ctx context.Context, req *types.DeleteRoleReq) (*types.DeleteRoleResp, error) {
	role, err := s.roleUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 检查数据权限
	if err := s.CheckRecordAccess(ctx, role.Role); err != nil {
		return nil, err
	}

	if err := s.roleUseCase.Delete(ctx, req.ID); err != nil {
		return nil, err
	}

	return &types.DeleteRoleResp{}, nil
}

func (s *RoleService) GetRoleOptions(ctx context.Context, req *types.GetRoleOptionsReq) (*types.GetRoleOptionsResp, error) {
	filters := map[string]any{}

	roles, err := s.roleUseCase.GetRoleOptions(ctx, filters, req.GetSortClause())
	if err != nil {
		return nil, err
	}

	results := make([]types.GetRoleOptionsRespResult, 0, len(roles))
	for _, role := range roles {
		results = append(results, types.GetRoleOptionsRespResult{
			GetOptionsRespResult: types.GetOptionsRespResult{
				Label: role.Name,
				Value: role.ID,
			},
			Slug:    role.Slug,
			Enabled: role.Enabled,
		})
	}

	return &types.GetRoleOptionsResp{
		Results: results,
	}, nil
}

// BatchDeleteRoles handles batch role deletion.
func (s *RoleService) BatchDeleteRoles(ctx context.Context, req *types.BatchDeleteRolesReq) (*types.BatchDeleteRolesResp, error) {
	// 先检查每个角色的数据权限
	accessibleIDs := make([]uint, 0, len(req.IDs))
	failedItems := make([]string, 0)

	for _, id := range req.IDs {
		role, err := s.roleUseCase.GetByID(ctx, id)
		if err != nil {
			// 角色不存在
			failedItems = append(failedItems, fmt.Sprintf("角色ID %d: 角色不存在或已被删除", id))
			continue
		}

		// 检查数据权限
		if err := s.CheckRecordAccess(ctx, role.Role); err != nil {
			failedItems = append(failedItems, role.Name+": 无权限删除此角色")
			continue
		}

		accessibleIDs = append(accessibleIDs, id)
	}

	// 批量删除有权限的角色
	results := s.roleUseCase.BatchDelete(ctx, accessibleIDs)

	// 统计结果
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			// 格式化错误信息：角色名称: 错误原因
			errMsg := result.Identifier + ": " + errs.GetMessage(result.Error)
			failedItems = append(failedItems, errMsg)
		}
	}

	return &types.BatchDeleteRolesResp{
		SuccessCount: successCount,
		FailedCount:  len(failedItems),
		FailedItems:  failedItems,
	}, nil
}
