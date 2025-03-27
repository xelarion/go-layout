// Package service provides service implementations that coordinate between handlers and usecases.
package service

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/usecase"
)

// UserRequest represents the user request data structure.
type UserRequest struct {
	ID       uint   `json:"id,omitempty"`
	Name     string `json:"name,omitempty" binding:"required,min=2,max=100"`
	Email    string `json:"email,omitempty" binding:"required,email"`
	Password string `json:"password,omitempty" binding:"required,min=6"`
	Role     string `json:"role,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
}

// UserResponse represents the user response data structure.
type UserResponse struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// LoginResponse represents the login response data structure.
type LoginResponse struct {
	User        UserResponse `json:"user"`
	Token       string       `json:"token"`
	TokenExpiry time.Time    `json:"token_expiry"`
}

// ListResponse represents the list response data structure.
type ListResponse struct {
	Users  []UserResponse `json:"users"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
	Count  int            `json:"count"`
}

// UserService coordinates between handlers and usecases for user operations.
type UserService struct {
	userUseCase *usecase.UserUseCase
	logger      *zap.Logger
}

// NewUserService creates a new instance of UserService.
func NewUserService(userUseCase *usecase.UserUseCase, logger *zap.Logger) *UserService {
	return &UserService{
		userUseCase: userUseCase,
		logger:      logger.Named("user_service"),
	}
}

// GetUser retrieves a user by ID.
func (s *UserService) GetUser(ctx context.Context, id uint) (*UserResponse, error) {
	user, err := s.userUseCase.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	if user == nil {
		return nil, nil
	}
	
	return &UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: &user.CreatedAt,
		UpdatedAt: &user.UpdatedAt,
	}, nil
}

// RegisterUser registers a new user.
func (s *UserService) RegisterUser(ctx context.Context, req *UserRequest) (*UserResponse, error) {
	user, err := s.userUseCase.Register(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	
	return &UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: &user.CreatedAt,
		UpdatedAt: &user.UpdatedAt,
	}, nil
}

// LoginUser authenticates a user.
func (s *UserService) LoginUser(ctx context.Context, req *UserRequest) (*UserResponse, error) {
	user, err := s.userUseCase.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

// UpdateUser updates user information.
func (s *UserService) UpdateUser(ctx context.Context, req *UserRequest) (*UserResponse, error) {
	user, err := s.userUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	
	if user == nil {
		return nil, usecase.ErrUserNotFound
	}
	
	// Update fields
	if req.Name != "" {
		user.Name = req.Name
	}
	
	if req.Email != "" {
		user.Email = req.Email
	}
	
	// Only admin can update role
	if req.Role != "" {
		user.Role = req.Role
	}
	
	if err := s.userUseCase.Update(ctx, user); err != nil {
		return nil, err
	}
	
	return &UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		UpdatedAt: &user.UpdatedAt,
	}, nil
}

// DeleteUser deletes a user.
func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	return s.userUseCase.Delete(ctx, id)
}

// ListUsers retrieves a list of users with pagination.
func (s *UserService) ListUsers(ctx context.Context, req *UserRequest) (*ListResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 10 // Default limit
	}
	
	users, err := s.userUseCase.List(ctx, limit, req.Offset)
	if err != nil {
		return nil, err
	}
	
	var responseUsers []UserResponse
	for _, user := range users {
		responseUsers = append(responseUsers, UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: &user.CreatedAt,
			UpdatedAt: &user.UpdatedAt,
		})
	}
	
	return &ListResponse{
		Users:  responseUsers,
		Limit:  limit,
		Offset: req.Offset,
		Count:  len(responseUsers),
	}, nil
}
