package service

import (
	"context"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	"github.com/xelarion/go-layout/internal/permission"
	"github.com/xelarion/go-layout/internal/usecase"
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

// nodeToResponse converts permission.Node to types.GetPermissionTreeRespPermission
func nodeToResponse(node *permission.Node) *types.GetPermissionTreeRespPermission {
	resp := &types.GetPermissionTreeRespPermission{
		ID:   node.ID,
		Name: node.Name,
	}

	if len(node.Children) > 0 {
		resp.Children = make([]*types.GetPermissionTreeRespPermission, 0, len(node.Children))
		for _, child := range node.Children {
			resp.Children = append(resp.Children, nodeToResponse(child))
		}
	}

	return resp
}

// GetPermissionTree returns the permission tree
func (s *PermissionService) GetPermissionTree(ctx context.Context, req *types.GetPermissionTreeReq) (*types.GetPermissionTreeResp, error) {
	// Get permission tree from usecase
	nodes, err := s.permissionUseCase.GetPermissionTree(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to response type
	respPerms := make([]*types.GetPermissionTreeRespPermission, 0, len(nodes))
	for _, node := range nodes {
		respPerms = append(respPerms, nodeToResponse(node))
	}

	return &types.GetPermissionTreeResp{
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

	return &types.GetRolePermissionsResp{
		Permissions: role.Permissions,
	}, nil
}

// UpdateRolePermissions updates the permissions for a role
func (s *PermissionService) UpdateRolePermissions(ctx context.Context, req *types.UpdateRolePermissionsReq) (*types.UpdateRolePermissionsResp, error) {
	// Update role permissions
	if err := s.roleUseCase.UpdatePermissions(ctx, req.RoleID, req.Permissions); err != nil {
		return nil, err
	}

	return &types.UpdateRolePermissionsResp{}, nil
}
