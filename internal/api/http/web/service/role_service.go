package service

import (
	"context"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	"github.com/xelarion/go-layout/internal/usecase"
)

// RoleService handles role-related services.
type RoleService struct {
	roleUseCase *usecase.RoleUseCase
}

// NewRoleService creates a new RoleService.
func NewRoleService(roleUseCase *usecase.RoleUseCase) *RoleService {
	return &RoleService{
		roleUseCase: roleUseCase,
	}
}

// CreateRole registers a new role.
func (s *RoleService) CreateRole(ctx context.Context, req *types.CreateRoleReq) (*types.CreateRoleResp, error) {
	params := usecase.CreateRoleParams{
		Name:        req.Name,
		Description: req.Description,
		Enabled:     true,
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
	if req.Name != "" {
		filters["name"] = req.Name
	}

	if req.Enabled != nil {
		filters["enabled"] = *req.Enabled
	}

	roles, count, err := s.roleUseCase.List(ctx, filters, req.GetLimit(), req.GetOffset(), req.GetSortClause())
	if err != nil {
		return nil, err
	}

	respResults := make([]types.ListRolesRespResult, 0, len(roles))
	for _, role := range roles {
		u := types.ListRolesRespResult{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			Enabled:     role.Enabled,
			CreatedAt:   types.Time(role.CreatedAt),
			UserCount:   role.UserCount,
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

	return &types.GetRoleResp{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Enabled:     role.Enabled,
		CreatedAt:   types.Time(role.CreatedAt),
	}, nil
}

// GetRoleFormData provides data needed for role forms (update).
func (s *RoleService) GetRoleFormData(ctx context.Context, req *types.GetRoleFormDataReq) (*types.GetRoleFormDataResp, error) {
	role, err := s.roleUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	return &types.GetRoleFormDataResp{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
	}, nil
}

// UpdateRole updates a role.
func (s *RoleService) UpdateRole(ctx context.Context, req *types.UpdateRoleReq) (*types.UpdateRoleResp, error) {
	// First check if the role exists
	_, err := s.roleUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// Create update params
	params := usecase.UpdateRoleParams{
		ID: req.ID,
	}

	params.Name = req.Name
	params.NameSet = true

	params.Description = req.Description
	params.DescriptionSet = true

	if err := s.roleUseCase.Update(ctx, params); err != nil {
		return nil, err
	}

	return &types.UpdateRoleResp{}, nil
}

// UpdateRoleEnabled updates a role's enabled status.
func (s *RoleService) UpdateRoleEnabled(ctx context.Context, req *types.UpdateRoleEnabledReq) (*types.UpdateRoleResp, error) {
	// First check if the role exists
	_, err := s.roleUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// Create update params with only the enabled status
	params := usecase.UpdateRoleParams{
		ID:         req.ID,
		Enabled:    *req.Enabled,
		EnabledSet: true,
	}

	if err := s.roleUseCase.Update(ctx, params); err != nil {
		return nil, err
	}

	return &types.UpdateRoleResp{}, nil
}

// DeleteRole handles role deletion.
func (s *RoleService) DeleteRole(ctx context.Context, req *types.DeleteRoleReq) (*types.DeleteRoleResp, error) {
	if err := s.roleUseCase.Delete(ctx, req.ID); err != nil {
		return nil, err
	}

	return &types.DeleteRoleResp{}, nil
}

func (s *RoleService) GetRoleOptions(ctx context.Context, req *types.GetRoleOptionsReq) (*types.GetRoleOptionsResp, error) {
	roles, err := s.roleUseCase.GetRoleOptions(ctx)
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
