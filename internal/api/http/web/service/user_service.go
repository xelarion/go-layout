package service

import (
	"context"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	"github.com/xelarion/go-layout/internal/usecase"
)

// UserService handles user-related services.
type UserService struct {
	userUseCase *usecase.UserUseCase
}

// NewUserService creates a new UserService.
func NewUserService(userUseCase *usecase.UserUseCase) *UserService {
	return &UserService{
		userUseCase: userUseCase,
	}
}

// CreateUser registers a new user.
func (s *UserService) CreateUser(ctx context.Context, req *types.CreateUserReq) (*types.CreateUserResp, error) {
	params := usecase.CreateUserParams{
		Username:     req.Username,
		Password:     req.Password,
		FullName:     req.FullName,
		PhoneNumber:  req.PhoneNumber,
		Email:        req.Email,
		Enabled:      true,
		DepartmentID: req.DepartmentID,
		RoleID:       req.RoleID,
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

	if req.DepartmentID != 0 {
		filters["department_id"] = req.DepartmentID
	}

	if req.Enabled != nil {
		filters["enabled"] = *req.Enabled
	}

	users, count, err := s.userUseCase.List(ctx, filters, req.GetLimit(), req.GetOffset(), req.GetSortClause())
	if err != nil {
		return nil, err
	}

	respResults := make([]types.ListUsersRespResult, 0, len(users))
	for _, user := range users {
		u := types.ListUsersRespResult{
			ID:             user.ID,
			CreatedAt:      types.Time(user.CreatedAt),
			Username:       user.Username,
			FullName:       user.FullName,
			PhoneNumber:    user.PhoneNumber,
			Email:          user.Email,
			RoleName:       user.RoleName,
			RoleSlug:       user.RoleSlug,
			Enabled:        user.Enabled,
			DepartmentName: user.DepartmentName,
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

	return &types.GetUserResp{
		ID:             user.ID,
		CreatedAt:      types.Time(user.CreatedAt),
		Username:       user.Username,
		FullName:       user.FullName,
		PhoneNumber:    user.PhoneNumber,
		Email:          user.Email,
		RoleName:       user.RoleName,
		RoleSlug:       user.RoleSlug,
		Enabled:        user.Enabled,
		DepartmentName: user.DepartmentName,
	}, nil
}

// GetUserFormData provides data needed for user forms (update).
func (s *UserService) GetUserFormData(ctx context.Context, req *types.GetUserFormDataReq) (*types.GetUserFormDataResp, error) {
	user, err := s.userUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	return &types.GetUserFormDataResp{
		ID:           user.ID,
		Username:     user.Username,
		FullName:     user.FullName,
		PhoneNumber:  user.PhoneNumber,
		Email:        user.Email,
		RoleID:       user.RoleID,
		DepartmentID: user.DepartmentID,
	}, nil
}

// UpdateUser updates a user.
func (s *UserService) UpdateUser(ctx context.Context, req *types.UpdateUserReq) (*types.UpdateUserResp, error) {
	// First check if the user exists
	_, err := s.userUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// Create update params
	params := usecase.UpdateUserParams{
		ID: req.ID,
	}

	params.Username = req.Username
	params.UsernameSet = true

	if req.Password != "" {
		params.Password = req.Password
		params.PasswordSet = true
	}

	params.FullName = req.FullName
	params.FullNameSet = true

	params.PhoneNumber = req.PhoneNumber
	params.PhoneNumberSet = true

	params.Email = req.Email
	params.EmailSet = true

	params.RoleID = req.RoleID
	params.RoleIDSet = true

	params.DepartmentID = req.DepartmentID
	params.DepartmentIDSet = true

	if err := s.userUseCase.Update(ctx, params); err != nil {
		return nil, err
	}

	return &types.UpdateUserResp{}, nil
}

// UpdateUserEnabled updates a user's enabled status.
func (s *UserService) UpdateUserEnabled(ctx context.Context, req *types.UpdateUserEnabledReq) (*types.UpdateUserResp, error) {
	// First check if the user exists
	_, err := s.userUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// Create update params with only the enabled status
	params := usecase.UpdateUserParams{
		ID:         req.ID,
		Enabled:    *req.Enabled,
		EnabledSet: true,
	}

	if err := s.userUseCase.Update(ctx, params); err != nil {
		return nil, err
	}

	return &types.UpdateUserResp{}, nil
}

// DeleteUser handles user deletion.
func (s *UserService) DeleteUser(ctx context.Context, req *types.DeleteUserReq) (*types.DeleteUserResp, error) {
	if err := s.userUseCase.Delete(ctx, req.ID); err != nil {
		return nil, err
	}

	return &types.DeleteUserResp{}, nil
}
