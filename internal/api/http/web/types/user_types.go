package types

import "time"

// Common user-related types

// CreateUserReq represents user creation request.
type CreateUserReq struct {
	Username string `json:"username" binding:"required,min=1,max=100"`
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
	PageReq
	SortReq
	Username string `form:"username" binding:"omitempty,max=100"`
	Email    string `form:"email" binding:"omitempty,max=100"`
	Role     string `form:"role" binding:"omitempty,oneof=user admin"`
	Enabled  *bool  `form:"enabled"`
}

// ListUsersResp represents user list with pagination info.
type ListUsersResp struct {
	Results  []ListUsersRespResult `json:"results"`
	PageInfo PageResp              `json:"page_info"`
}

type ListUsersRespResult struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Enabled   bool      `json:"enabled"`
	Avatar    string    `json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
}

// GetUserReq represents a request to get a specific user.
type GetUserReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// GetUserResp represents a user object in responses.
type GetUserResp struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Enabled   bool      `json:"enabled"`
	Avatar    string    `json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
}

type GetUserFormDataReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// GetUserFormDataResp represents data needed for user forms (create/update).
type GetUserFormDataResp struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// UpdateUserReq represents user update data.
type UpdateUserReq struct {
	ID       uint   `uri:"id" binding:"required" swaggerignore:"true"`
	Username string `json:"username" binding:"required,min=1,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"omitempty,min=6"`
	Role     string `json:"role" binding:"required,oneof=user admin"`
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
