package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	appctx "github.com/xelarion/go-layout/internal/context"
	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/errs"
)

// OrganizationService handles organization-related services.
type OrganizationService struct {
	*BaseService        // 嵌入 BaseService 提供数据权限检查方法
	organizationUseCase *usecase.OrganizationUseCase
	logger              *zap.Logger
}

// NewOrganizationService creates a new OrganizationService.
func NewOrganizationService(
	baseService *BaseService,
	organizationUseCase *usecase.OrganizationUseCase,
	logger *zap.Logger,
) *OrganizationService {
	return &OrganizationService{
		BaseService:         baseService,
		organizationUseCase: organizationUseCase,
		logger:              logger,
	}
}

// CreateOrganization registers a new organization.
func (s *OrganizationService) CreateOrganization(ctx context.Context, req *types.CreateOrganizationReq) (*types.CreateOrganizationResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	orgType, err := enum.GetOrganizationType(req.Type)
	if err != nil {
		return nil, err
	}

	params := usecase.CreateOrganizationParams{
		CreateRecordCommonParams: usecase.CreateRecordCommonParams{
			CreatedBy:     current.UserID,
			CreatedRoleID: current.RoleID,
			AssignedTo:    0, // organizations 表不使用 assigned_to
		},
		Code:     req.Code,
		Name:     req.Name,
		ParentID: req.ParentID,
		Type:     orgType,
		OwnerID:  req.OwnerID,
		Remark:   req.Remark,
	}

	id, err := s.organizationUseCase.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	return &types.CreateOrganizationResp{
		ID: id,
	}, nil
}

// ListOrganizations lists organizations with pagination.
func (s *OrganizationService) ListOrganizations(ctx context.Context, req *types.ListOrganizationsReq) (*types.ListOrganizationsResp, error) {
	filters := map[string]any{}
	if req.Key != "" {
		filters["key"] = req.Key
	}

	if req.Type != "" {
		orgType, err := enum.GetOrganizationType(req.Type)
		if err != nil {
			return nil, err
		}
		filters["type"] = orgType
	}

	if req.OwnerID != 0 {
		filters["owner_id"] = req.OwnerID
	}

	// 应用数据权限过滤（传入空的 Organization 实例用于获取字段映射）
	if err := s.ApplyDataScope(ctx, filters, &model.Organization{}); err != nil {
		return nil, err
	}

	organizations, _, err := s.organizationUseCase.List(ctx, filters, -1, -1, req.GetSortClause())
	if err != nil {
		return nil, err
	}

	return &types.ListOrganizationsResp{
		Results: s.buildListOrganizationsTree(organizations),
	}, nil
}

func (s *OrganizationService) buildListOrganizationsTree(organizations []*usecase.Organization) []*types.ListOrganizationsRespResult {
	// Create a map for quick lookup
	nodeMap := make(map[uint]*types.ListOrganizationsRespResult)
	var rootNodes []*types.ListOrganizationsRespResult

	// First pass: create all nodes
	for _, organization := range organizations {
		node := &types.ListOrganizationsRespResult{
			ID:        organization.ID,
			Name:      organization.Name,
			Code:      organization.Code,
			Type:      enum.OrganizationTypeText[organization.Type],
			ParentID:  organization.ParentID,
			Owner:     organization.OwnerName,
			UserCount: organization.UserCount,
			Remark:    organization.Remark,
			Children:  make([]*types.ListOrganizationsRespResult, 0),
		}
		nodeMap[organization.ID] = node
	}

	// Second pass: build tree structure
	for _, org := range organizations {
		node := nodeMap[org.ID]
		if org.ParentID == nil {
			// Root node (no parent)
			rootNodes = append(rootNodes, node)
		} else {
			// Child node, add to parent
			if parent, exists := nodeMap[*org.ParentID]; exists {
				parent.Children = append(parent.Children, node)
			} else {
				// Parent not found, treat as root
				rootNodes = append(rootNodes, node)
			}
		}
	}

	// Clean up empty children slices
	for _, node := range nodeMap {
		if len(node.Children) == 0 {
			node.Children = nil
		}
	}

	return rootNodes
}

// GetOrganizationFormData provides data needed for organization forms (update).
func (s *OrganizationService) GetOrganizationFormData(ctx context.Context, req *types.GetOrganizationFormDataReq) (*types.GetOrganizationFormDataResp, error) {
	organization, err := s.organizationUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 检查数据权限
	if err := s.CheckRecordAccess(ctx, organization.Organization); err != nil {
		return nil, err
	}

	return &types.GetOrganizationFormDataResp{
		ID:       organization.ID,
		Code:     organization.Code,
		Name:     organization.Name,
		ParentID: organization.ParentID,
		Type:     enum.OrganizationTypeSlug[organization.Type],
		OwnerID:  organization.OwnerID,
		Remark:   organization.Remark,
	}, nil
}

// UpdateOrganization updates a organization.
func (s *OrganizationService) UpdateOrganization(ctx context.Context, req *types.UpdateOrganizationReq) (*types.UpdateOrganizationResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	// 先获取角色，用于权限检查
	organization, err := s.organizationUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 检查数据权限
	if err := s.CheckRecordAccess(ctx, organization.Organization); err != nil {
		return nil, err
	}

	// Create update params
	params := usecase.UpdateOrganizationParams{
		UpdateRecordCommonParams: usecase.UpdateRecordCommonParams{
			UpdatedBy: current.UserID,
		},
		ID: req.ID,
	}

	params.Code = req.Code
	params.CodeSet = true

	params.Name = req.Name
	params.NameSet = true

	params.ParentID = req.ParentID
	params.ParentIDSet = true

	params.Type, err = enum.GetOrganizationType(req.Type)
	if err != nil {
		return nil, err
	}
	params.TypeSet = true

	params.OwnerID = req.OwnerID
	params.OwnerIDSet = true

	params.Remark = req.Remark
	params.RemarkSet = true

	if err := s.organizationUseCase.Update(ctx, params); err != nil {
		return nil, err
	}

	return &types.UpdateOrganizationResp{}, nil
}

// DeleteOrganization handles organization deletion.
func (s *OrganizationService) DeleteOrganization(ctx context.Context, req *types.DeleteOrganizationReq) (*types.DeleteOrganizationResp, error) {
	organization, err := s.organizationUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 检查数据权限
	if err := s.CheckRecordAccess(ctx, organization.Organization); err != nil {
		return nil, err
	}

	if err := s.organizationUseCase.Delete(ctx, req.ID); err != nil {
		return nil, err
	}

	return &types.DeleteOrganizationResp{}, nil
}

// GetOrganizationsTree returns the organization tree
func (s *OrganizationService) GetOrganizationsTree(ctx context.Context, req *types.GetOrganizationsTreeReq) (*types.GetOrganizationsTreeResp, error) {
	organizations, err := s.organizationUseCase.GetAllOrganizations(ctx)
	if err != nil {
		return nil, err
	}

	// Build tree structure from flat list
	tree := s.buildOrganizationTree(organizations)

	return &types.GetOrganizationsTreeResp{
		Results: tree,
	}, nil
}

// buildOrganizationTree builds a tree structure from flat organization list
func (s *OrganizationService) buildOrganizationTree(organizations []*model.Organization) []*types.TreeNodeResp {
	// Create a map for quick lookup
	nodeMap := make(map[uint]*types.TreeNodeResp)
	var rootNodes []*types.TreeNodeResp

	// First pass: create all nodes
	for _, org := range organizations {
		node := &types.TreeNodeResp{
			Key:      org.ID,
			Title:    org.Name,
			Children: make([]*types.TreeNodeResp, 0),
		}
		nodeMap[org.ID] = node
	}

	// Second pass: build tree structure
	for _, org := range organizations {
		node := nodeMap[org.ID]
		if org.ParentID == nil {
			// Root node (no parent)
			rootNodes = append(rootNodes, node)
		} else {
			// Child node, add to parent
			if parent, exists := nodeMap[*org.ParentID]; exists {
				parent.Children = append(parent.Children, node)
			} else {
				// Parent not found, treat as root
				rootNodes = append(rootNodes, node)
			}
		}
	}

	// Clean up empty children slices
	for _, node := range nodeMap {
		if len(node.Children) == 0 {
			node.Children = nil
		}
	}

	return rootNodes
}
