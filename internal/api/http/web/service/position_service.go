package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	appctx "github.com/xelarion/go-layout/internal/context"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/errs"
)

// PositionService handles position-related services.
type PositionService struct {
	*BaseService    // 嵌入 BaseService 提供数据权限检查方法
	positionUseCase *usecase.PositionUseCase
	logger          *zap.Logger
}

// NewPositionService creates a new PositionService.
func NewPositionService(
	baseService *BaseService,
	positionUseCase *usecase.PositionUseCase,
	logger *zap.Logger,
) *PositionService {
	return &PositionService{
		BaseService:     baseService,
		positionUseCase: positionUseCase,
		logger:          logger,
	}
}

// CreatePosition registers a new position.
func (s *PositionService) CreatePosition(ctx context.Context, req *types.CreatePositionReq) (*types.CreatePositionResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	params := usecase.CreatePositionParams{
		CreateRecordCommonParams: usecase.CreateRecordCommonParams{
			CreatedBy:     current.UserID,
			CreatedRoleID: current.RoleID,
			AssignedTo:    0, // positions 表不使用 assigned_to
		},
		Code:           req.Code,
		Name:           req.Name,
		OrganizationID: req.OrganizationID,
		Remark:         req.Remark,
	}

	id, err := s.positionUseCase.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	return &types.CreatePositionResp{
		ID: id,
	}, nil
}

// ListPositions lists positions with pagination.
func (s *PositionService) ListPositions(ctx context.Context, req *types.ListPositionsReq) (*types.ListPositionsResp, error) {
	filters := map[string]any{}
	if req.Key != "" {
		filters["key"] = req.Key
	}

	if req.OrganizationID > 0 {
		filters["organization_id"] = req.OrganizationID
	}

	// 应用数据权限过滤（传入空的 Position 实例用于获取字段映射）
	if err := s.ApplyDataScope(ctx, filters, &model.Position{}); err != nil {
		return nil, err
	}

	positions, count, err := s.positionUseCase.List(ctx, filters, req.GetLimit(), req.GetOffset(), req.GetSortClause())
	if err != nil {
		return nil, err
	}

	respResults := make([]types.ListPositionsRespResult, 0, len(positions))
	for _, position := range positions {
		u := types.ListPositionsRespResult{
			ID:               position.ID,
			Name:             position.Name,
			Code:             position.Code,
			OrganizationName: position.OrganizationName,
			UserCount:        position.UserCount,
			Remark:           position.Remark,
		}
		respResults = append(respResults, u)
	}

	return &types.ListPositionsResp{
		Results:    respResults,
		Pagination: types.NewPageResp(count, req.GetPage(), req.GetPageSize()),
	}, nil
}

// GetPositionFormData provides data needed for position forms (update).
func (s *PositionService) GetPositionFormData(ctx context.Context, req *types.GetPositionFormDataReq) (*types.GetPositionFormDataResp, error) {
	position, err := s.positionUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 检查数据权限
	if err := s.CheckRecordAccess(ctx, position.Position); err != nil {
		return nil, err
	}

	return &types.GetPositionFormDataResp{
		ID:             position.ID,
		Code:           position.Code,
		Name:           position.Name,
		OrganizationID: position.OrganizationID,
		Remark:         position.Remark,
	}, nil
}

// UpdatePosition updates a position.
func (s *PositionService) UpdatePosition(ctx context.Context, req *types.UpdatePositionReq) (*types.UpdatePositionResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	position, err := s.positionUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 检查数据权限
	if err := s.CheckRecordAccess(ctx, position.Position); err != nil {
		return nil, err
	}

	// Create update params
	params := usecase.UpdatePositionParams{
		UpdateRecordCommonParams: usecase.UpdateRecordCommonParams{
			UpdatedBy: current.UserID,
		},
		ID: req.ID,
	}

	params.Code = req.Code
	params.CodeSet = true

	params.Name = req.Name
	params.NameSet = true

	params.OrganizationID = req.OrganizationID
	params.OrganizationIDSet = true

	params.Remark = req.Remark
	params.RemarkSet = true

	if err := s.positionUseCase.Update(ctx, params); err != nil {
		return nil, err
	}

	return &types.UpdatePositionResp{}, nil
}

// DeletePosition handles position deletion.
func (s *PositionService) DeletePosition(ctx context.Context, req *types.DeletePositionReq) (*types.DeletePositionResp, error) {
	position, err := s.positionUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 检查数据权限
	if err := s.CheckRecordAccess(ctx, position.Position); err != nil {
		return nil, err
	}

	if err := s.positionUseCase.Delete(ctx, req.ID); err != nil {
		return nil, err
	}

	return &types.DeletePositionResp{}, nil
}

func (s *PositionService) GetPositionOptions(ctx context.Context, req *types.GetPositionOptionsReq) (*types.GetPositionOptionsResp, error) {
	filters := map[string]any{}

	positions, err := s.positionUseCase.GetPositionOptions(ctx, filters, req.GetSortClause())
	if err != nil {
		return nil, err
	}

	results := make([]types.GetPositionOptionsRespResult, 0, len(positions))
	for _, position := range positions {
		results = append(results, types.GetPositionOptionsRespResult{
			GetOptionsRespResult: types.GetOptionsRespResult{
				Label: position.Name,
				Value: position.ID,
			},
			OrganizationName: position.OrganizationName,
		})
	}

	return &types.GetPositionOptionsResp{
		Results: results,
	}, nil
}

// BatchDeletePositions handles batch position deletion.
func (s *PositionService) BatchDeletePositions(ctx context.Context, req *types.BatchDeletePositionsReq) (*types.BatchDeletePositionsResp, error) {
	// 先检查每个角色的数据权限
	accessibleIDs := make([]uint, 0, len(req.IDs))
	failedItems := make([]string, 0)

	for _, id := range req.IDs {
		position, err := s.positionUseCase.GetByID(ctx, id)
		if err != nil {
			// 角色不存在
			failedItems = append(failedItems, fmt.Sprintf("岗位ID %d: 岗位不存在或已被删除", id))
			continue
		}

		// 检查数据权限
		if err := s.CheckRecordAccess(ctx, position.Position); err != nil {
			failedItems = append(failedItems, position.Name+": 无权限删除此岗位")
			continue
		}

		accessibleIDs = append(accessibleIDs, id)
	}

	// 批量删除有权限的角色
	results := s.positionUseCase.BatchDelete(ctx, accessibleIDs)

	// 统计结果
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			// 格式化错误信息：岗位名称: 错误原因
			errMsg := result.Identifier + ": " + errs.GetMessage(result.Error)
			failedItems = append(failedItems, errMsg)
		}
	}

	return &types.BatchDeletePositionsResp{
		SuccessCount: successCount,
		FailedCount:  len(failedItems),
		FailedItems:  failedItems,
	}, nil
}
