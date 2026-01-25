package service

import (
	"context"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	appctx "github.com/xelarion/go-layout/internal/context"
	"github.com/xelarion/go-layout/internal/permission"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/errs"
)

// PermissionService handles permission related business logic
type PermissionService struct {
	permissionUseCase *usecase.PermissionUseCase
	roleUseCase       *usecase.RoleUseCase
}

// NewPermissionService creates a new instance of PermissionService
func NewPermissionService(
	permissionUseCase *usecase.PermissionUseCase,
	roleUseCase *usecase.RoleUseCase,
) *PermissionService {
	return &PermissionService{
		permissionUseCase: permissionUseCase,
		roleUseCase:       roleUseCase,
	}
}

// permissionNodeToTreeNodeResp converts permission.Node to types.TreeNodeResp
func (s *PermissionService) permissionNodeToTreeNodeResp(node *permission.Node) *types.TreeNodeResp {
	resp := &types.TreeNodeResp{
		Key:   node.ID,
		Title: node.Name,
	}

	if len(node.Children) > 0 {
		resp.Children = make([]*types.TreeNodeResp, 0, len(node.Children))
		for _, child := range node.Children {
			resp.Children = append(resp.Children, s.permissionNodeToTreeNodeResp(child))
		}
	}

	return resp
}

// GetPermissionsTree returns the permission tree
func (s *PermissionService) GetPermissionsTree(ctx context.Context, req *types.GetPermissionsTreeReq) (*types.GetPermissionsTreeResp, error) {
	// Get permission tree from usecase
	nodes, err := s.permissionUseCase.GetPermissionsTree(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to response type
	respPerms := make([]*types.TreeNodeResp, 0, len(nodes))
	for _, node := range nodes {
		respPerms = append(respPerms, s.permissionNodeToTreeNodeResp(node))
	}

	return &types.GetPermissionsTreeResp{
		Permissions: respPerms,
	}, nil
}

// GetRolePermissions returns the permissions for a role
func (s *PermissionService) GetRolePermissions(ctx context.Context, req *types.GetRolePermissionsReq) (*types.GetRolePermissionsResp, error) {
	// Get role from usecase
	role, err := s.roleUseCase.GetByID(ctx, req.RoleID)
	if err != nil {
		return nil, err
	}

	dataScopes := make([]types.GetRolePermissionsRespDataScope, 0, len(role.DataScopes))
	for _, dataScope := range role.DataScopes {
		dataScopes = append(dataScopes, types.GetRolePermissionsRespDataScope{
			Permissions: dataScope.Permissions,
			Type:        string(dataScope.Type),
		})
	}

	return &types.GetRolePermissionsResp{
		Permissions: role.Permissions,
		DataScopes:  dataScopes,
	}, nil
}

// UpdateRolePermissions updates the permissions for a role
func (s *PermissionService) UpdateRolePermissions(ctx context.Context, req *types.UpdateRolePermissionsReq) (*types.UpdateRolePermissionsResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	dataScopes := make([]usecase.UpdatePermissionsDataScope, 0, len(req.DataScopes))
	for _, dataScope := range req.DataScopes {
		dataScopes = append(dataScopes, usecase.UpdatePermissionsDataScope{
			Permissions: dataScope.Permissions,
			Type:        dataScope.Type,
		})
	}
	params := usecase.UpdatePermissionsParams{
		UpdateRecordCommonParams: usecase.UpdateRecordCommonParams{
			UpdatedBy: current.UserID,
		},
		ID:          req.RoleID,
		Permissions: req.Permissions,
		DataScopes:  dataScopes,
	}

	if err := s.roleUseCase.UpdatePermissions(ctx, params); err != nil {
		return nil, err
	}

	return &types.UpdateRolePermissionsResp{}, nil
}
