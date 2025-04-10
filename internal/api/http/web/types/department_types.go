package types

import "time"

// Common department-related types

// CreateDepartmentReq represents department creation request.
type CreateDepartmentReq struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"required"`
	Enabled     bool   `json:"enabled"`
}

// CreateDepartmentResp represents department creation response.
type CreateDepartmentResp struct {
	ID uint `json:"id"`
}

// ListDepartmentsReq represents department list query parameters.
type ListDepartmentsReq struct {
	PageReq
	SortReq
	Name    string `form:"name" binding:"omitempty,max=100"`
	Enabled *bool  `form:"enabled"`
}

// ListDepartmentsResp represents department list with pagination info.
type ListDepartmentsResp struct {
	Results    []ListDepartmentsRespResult `json:"results"`
	Pagination PageResp                    `json:"pagination"`
}

type ListDepartmentsRespResult struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
}

// GetDepartmentReq represents a request to get a specific department.
type GetDepartmentReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// GetDepartmentResp represents a department object in responses.
type GetDepartmentResp struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
}

type GetDepartmentFormDataReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// GetDepartmentFormDataResp represents data needed for department forms (create/update).
type GetDepartmentFormDataResp struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateDepartmentReq represents department update data.
type UpdateDepartmentReq struct {
	ID          uint   `uri:"id" binding:"required" swaggerignore:"true"`
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"required"`
	Enabled     bool   `json:"enabled" binding:"required"`
}

// UpdateDepartmentResp represents department update response.
type UpdateDepartmentResp struct {
}

// UpdateDepartmentEnabledReq represents a request to update department enabled status.
type UpdateDepartmentEnabledReq struct {
	ID      uint  `uri:"id" binding:"required" swaggerignore:"true"`
	Enabled *bool `json:"enabled" binding:"required"`
}

// UpdateDepartmentEnabledResp represents department update enabled status response.
type UpdateDepartmentEnabledResp struct {
}

// DeleteDepartmentReq represents a request to delete a department.
type DeleteDepartmentReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// DeleteDepartmentResp represents department delete response.
type DeleteDepartmentResp struct {
}
