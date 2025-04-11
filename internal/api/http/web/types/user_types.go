package types

import "time"

// Common user-related types

// CreateUserReq represents user creation request.
type CreateUserReq struct {
	Username     string `json:"username" binding:"required,min=1,max=100"`
	Password     string `json:"password" binding:"required,min=6"`
	FullName     string `json:"full_name" binding:"required,min=1,max=100"`
	PhoneNumber  string `json:"phone_number" binding:"required,min=1,max=20"`
	Email        string `json:"email" binding:"omitempty,email,max=100"`
	DepartmentID uint   `json:"department_id" binding:"required,min=1"`
	RoleID       uint   `json:"role_id" binding:"required,min=1"`
}

// CreateUserResp represents user creation response.
type CreateUserResp struct {
	ID uint `json:"id"`
}

// ListUsersReq represents user list query parameters.
type ListUsersReq struct {
	PageReq
	SortReq
	Key          string `form:"key" binding:"omitempty,max=100"` // Search by username or full name
	Enabled      *bool  `form:"enabled"`
	RoleID       uint   `form:"role_id" binding:"omitempty,gte=1"`
	DepartmentID uint   `form:"department_id" binding:"omitempty,gte=1"`
}

// ListUsersResp represents user list with pagination info.
type ListUsersResp struct {
	Results    []ListUsersRespResult `json:"results"`
	Pagination PageResp              `json:"pagination"`
}

type ListUsersRespResult struct {
	ID             uint      `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	Username       string    `json:"username"`
	FullName       string    `json:"full_name"`
	PhoneNumber    string    `json:"phone_number"`
	Email          string    `json:"email"`
	RoleName       string    `json:"role_name"`
	RoleSlug       string    `json:"role_slug"`
	Enabled        bool      `json:"enabled"`
	DepartmentName string    `json:"department_name"`
}

// GetUserReq represents a request to get a specific user.
type GetUserReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// GetUserResp represents a user object in responses.
type GetUserResp struct {
	ID             uint      `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	Username       string    `json:"username"`
	FullName       string    `json:"full_name"`
	PhoneNumber    string    `json:"phone_number"`
	Email          string    `json:"email"`
	RoleName       string    `json:"role_name"`
	RoleSlug       string    `json:"role_slug"`
	Enabled        bool      `json:"enabled"`
	DepartmentName string    `json:"department_name"`
}

type GetUserFormDataReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// GetUserFormDataResp represents data needed for user forms (create/update).
type GetUserFormDataResp struct {
	ID           uint   `json:"id"`
	Username     string `json:"username"`
	FullName     string `json:"full_name"`
	PhoneNumber  string `json:"phone_number"`
	Email        string `json:"email"`
	RoleID       uint   `json:"role_id"`
	DepartmentID uint   `json:"department_id"`
}

// UpdateUserReq represents user update data.
type UpdateUserReq struct {
	ID           uint   `uri:"id" binding:"required" swaggerignore:"true"`
	Username     string `json:"username" binding:"required,min=1,max=100"`
	Password     string `json:"password" binding:"omitempty,min=6"`
	FullName     string `json:"full_name" binding:"required,min=1,max=100"`
	PhoneNumber  string `json:"phone_number" binding:"required,min=1,max=20"`
	Email        string `json:"email" binding:"omitempty,email,max=100"`
	RoleID       uint   `json:"role_id" binding:"required,gte=1"`
	DepartmentID uint   `json:"department_id" binding:"required,gte=1"`
}

// UpdateUserResp represents user update response.
type UpdateUserResp struct {
}

// UpdateUserEnabledReq represents a request to update user enabled status.
type UpdateUserEnabledReq struct {
	ID      uint  `uri:"id" binding:"required" swaggerignore:"true"`
	Enabled *bool `json:"enabled" binding:"required"`
}

// UpdateUserEnabledResp represents user update enabled status response.
type UpdateUserEnabledResp struct {
}

// DeleteUserReq represents a request to delete a user.
type DeleteUserReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// DeleteUserResp represents user delete response.
type DeleteUserResp struct {
}
