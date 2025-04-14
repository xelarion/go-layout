package types

// Common role-related types

// CreateRoleReq represents role creation request.
type CreateRoleReq struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description"`
}

// CreateRoleResp represents role creation response.
type CreateRoleResp struct {
	ID uint `json:"id"`
}

// ListRolesReq represents role list query parameters.
type ListRolesReq struct {
	PageReq
	SortReq
	Name    string `form:"name" binding:"omitempty,max=100"`
	Enabled *bool  `form:"enabled"`
}

// ListRolesResp represents role list with pagination info.
type ListRolesResp struct {
	Results    []ListRolesRespResult `json:"results"`
	Pagination PageResp              `json:"pagination"`
}

// ListRolesRespResult represents a single role in the list.
type ListRolesRespResult struct {
	ID          uint   `json:"id"`
	CreatedAt   Time   `json:"created_at"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	UserCount   int64  `json:"user_count"`
}

// GetRoleReq represents a request to get a specific role.
type GetRoleReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// GetRoleResp represents a role object in responses.
type GetRoleResp struct {
	ID          uint   `json:"id"`
	CreatedAt   Time   `json:"created_at"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	UserCount   int64  `json:"user_count"`
}

type GetRoleFormDataReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// GetRoleFormDataResp represents data needed for role forms (create/update).
type GetRoleFormDataResp struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateRoleReq represents role update data.
type UpdateRoleReq struct {
	ID          uint   `uri:"id" binding:"required" swaggerignore:"true"`
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description"`
}

// UpdateRoleResp represents role update response.
type UpdateRoleResp struct {
}

// UpdateRoleEnabledReq represents a request to update role enabled status.
type UpdateRoleEnabledReq struct {
	ID      uint  `uri:"id" binding:"required" swaggerignore:"true"`
	Enabled *bool `json:"enabled" binding:"required"`
}

// UpdateRoleEnabledResp represents role update enabled status response.
type UpdateRoleEnabledResp struct {
}

// DeleteRoleReq represents a request to delete a role.
type DeleteRoleReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// DeleteRoleResp represents role delete response.
type DeleteRoleResp struct {
}

// GetRoleOptionsReq represents a request to get role options.
type GetRoleOptionsReq struct {
}

// GetRoleOptionsResp represents role options response.
type GetRoleOptionsResp struct {
	Results []GetRoleOptionsRespResult `json:"results"`
}
type GetRoleOptionsRespResult struct {
	GetOptionsRespResult
	Slug    string `json:"slug"`
	Enabled bool   `json:"enabled"`
}
