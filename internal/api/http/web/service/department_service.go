package service

import (
	"context"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	"github.com/xelarion/go-layout/internal/usecase"
)

// DepartmentService handles department-related services.
type DepartmentService struct {
	departmentUseCase *usecase.DepartmentUseCase
}

// NewDepartmentService creates a new DepartmentService.
func NewDepartmentService(departmentUseCase *usecase.DepartmentUseCase) *DepartmentService {
	return &DepartmentService{
		departmentUseCase: departmentUseCase,
	}
}

// CreateDepartment registers a new department.
func (s *DepartmentService) CreateDepartment(ctx context.Context, req *types.CreateDepartmentReq) (*types.CreateDepartmentResp, error) {
	params := usecase.CreateDepartmentParams{
		Name:        req.Name,
		Description: req.Description,
		Enabled:     true,
	}

	id, err := s.departmentUseCase.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	return &types.CreateDepartmentResp{
		ID: id,
	}, nil
}

// ListDepartments lists departments with pagination.
func (s *DepartmentService) ListDepartments(ctx context.Context, req *types.ListDepartmentsReq) (*types.ListDepartmentsResp, error) {
	filters := map[string]any{}
	if req.Name != "" {
		filters["name"] = req.Name
	}

	if req.Enabled != nil {
		filters["enabled"] = *req.Enabled
	}

	departments, count, err := s.departmentUseCase.List(ctx, filters, req.GetLimit(), req.GetOffset(), req.GetSortClause())
	if err != nil {
		return nil, err
	}

	respResults := make([]types.ListDepartmentsRespResult, 0, len(departments))
	for _, department := range departments {
		u := types.ListDepartmentsRespResult{
			ID:          department.ID,
			Name:        department.Name,
			Description: department.Description,
			Enabled:     department.Enabled,
			CreatedAt:   types.Time(department.CreatedAt),
			UserCount:   department.UserCount,
		}
		respResults = append(respResults, u)
	}

	return &types.ListDepartmentsResp{
		Results:    respResults,
		Pagination: types.NewPageResp(count, req.GetPage(), req.GetPageSize()),
	}, nil
}

// GetDepartment retrieves a department by ID.
func (s *DepartmentService) GetDepartment(ctx context.Context, req *types.GetDepartmentReq) (*types.GetDepartmentResp, error) {
	department, err := s.departmentUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	return &types.GetDepartmentResp{
		ID:          department.ID,
		Name:        department.Name,
		Description: department.Description,
		Enabled:     department.Enabled,
		CreatedAt:   types.Time(department.CreatedAt),
	}, nil
}

// GetDepartmentFormData provides data needed for department forms (update).
func (s *DepartmentService) GetDepartmentFormData(ctx context.Context, req *types.GetDepartmentFormDataReq) (*types.GetDepartmentFormDataResp, error) {
	department, err := s.departmentUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	return &types.GetDepartmentFormDataResp{
		ID:          department.ID,
		Name:        department.Name,
		Description: department.Description,
	}, nil
}

// UpdateDepartment updates a department.
func (s *DepartmentService) UpdateDepartment(ctx context.Context, req *types.UpdateDepartmentReq) (*types.UpdateDepartmentResp, error) {
	// Create update params
	params := usecase.UpdateDepartmentParams{
		ID: req.ID,
	}

	params.Name = req.Name
	params.NameSet = true

	params.Description = req.Description
	params.DescriptionSet = true

	if err := s.departmentUseCase.Update(ctx, params); err != nil {
		return nil, err
	}

	return &types.UpdateDepartmentResp{}, nil
}

// UpdateDepartmentEnabled updates a department's enabled status.
func (s *DepartmentService) UpdateDepartmentEnabled(ctx context.Context, req *types.UpdateDepartmentEnabledReq) (*types.UpdateDepartmentResp, error) {
	if err := s.departmentUseCase.UpdateEnabled(ctx, req.ID, *req.Enabled); err != nil {
		return nil, err
	}

	return &types.UpdateDepartmentResp{}, nil
}

// DeleteDepartment handles department deletion.
func (s *DepartmentService) DeleteDepartment(ctx context.Context, req *types.DeleteDepartmentReq) (*types.DeleteDepartmentResp, error) {
	if err := s.departmentUseCase.Delete(ctx, req.ID); err != nil {
		return nil, err
	}

	return &types.DeleteDepartmentResp{}, nil
}

func (s *DepartmentService) GetDepartmentOptions(ctx context.Context, req *types.GetDepartmentOptionsReq) (*types.GetDepartmentOptionsResp, error) {
	departments, err := s.departmentUseCase.GetDepartmentOptions(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]types.GetDepartmentOptionsRespResult, 0, len(departments))
	for _, department := range departments {
		results = append(results, types.GetDepartmentOptionsRespResult{
			GetOptionsRespResult: types.GetOptionsRespResult{
				Label: department.Name,
				Value: department.ID,
			},
			Enabled: department.Enabled,
		})
	}

	return &types.GetDepartmentOptionsResp{
		Results: results,
	}, nil
}
