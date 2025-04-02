// Package types contains request and response types for the web API.
package types

import "time"

// Common user-related types

// CreateUserReq represents user creation request.
type CreateUserReq struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=user admin"`
}

// CreateUserResp represents user creation response.
type CreateUserResp struct {
	ID uint `json:"id"`
}

// ListUsersReq represents user list query parameters.
type ListUsersReq struct {
	PageReq        // embed PageReq struct from common.go
	Name    string `form:"name" binding:"omitempty,max=100"`
	Email   string `form:"email" binding:"omitempty,max=100"`
	Role    string `form:"role" binding:"omitempty,oneof=user admin"`
	Enabled *bool  `form:"enabled"`
}

// ListUsersResp represents user list with pagination info.
type ListUsersResp struct {
	Results  []ListUsersRespResult `json:"results"`
	PageInfo PageInfoResp          `json:"page_info"`
}

type ListUsersRespResult struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	Enabled   bool       `json:"enabled"`
	Avatar    string     `json:"avatar"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

// GetUserReq represents a request to get a specific user.
type GetUserReq struct {
	ID uint `uri:"id" binding:"required"`
}

// GetUserResp represents a user object in responses.
type GetUserResp struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	Enabled   bool       `json:"enabled"`
	Avatar    string     `json:"avatar"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type GetUserFormDataReq struct {
	ID uint `uri:"id" binding:"required"`
}

// GetUserFormDataResp represents data needed for user forms (create/update).
type GetUserFormDataResp struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// UpdateUserReq represents user update data.
type UpdateUserReq struct {
	ID       uint   `uri:"id" binding:"required"`
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"omitempty,min=6"`
	Role     string `json:"role" binding:"required,oneof=user admin"`
}

// UpdateUserResp represents user update response.
type UpdateUserResp struct {
}

// UpdateUserEnabledReq represents a request to update user enabled status.
type UpdateUserEnabledReq struct {
	ID      uint  `uri:"id" binding:"required"`
	Enabled *bool `json:"enabled" binding:"required"`
}

// UpdateUserEnabledResp represents user update enabled status response.
type UpdateUserEnabledResp struct {
}

// DeleteUserReq represents a request to delete a user.
type DeleteUserReq struct {
	ID uint `uri:"id" binding:"required"`
}

// DeleteUserResp represents user delete response.
type DeleteUserResp struct {
}

type GetProfileReq struct {
}

type GetProfileResp struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	Avatar    string     `json:"avatar"`
	CreatedAt *time.Time `json:"created_at"`
}

// UpdateProfileReq represents profile update data.
type UpdateProfileReq struct {
	Name     string `json:"name" binding:"omitempty,min=2,max=100"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"omitempty,min=6"`
	Avatar   string `json:"avatar" binding:"omitempty,url"`
}

// UpdateProfileResp represents profile update response.
type UpdateProfileResp struct {
}
