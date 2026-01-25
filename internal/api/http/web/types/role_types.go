package types

// Common role-related types

// CreateRoleReq represents role creation request.
type CreateRoleReq struct {
	Code         string `json:"code" binding:"required,min=1,max=100"` // 角色编号
	Name         string `json:"name" binding:"required,min=1,max=100"` // 角色名称
	DisplayOrder int32  `json:"display_order" binding:"required"`      // 显示顺序
	Enabled      bool   `json:"enabled"`                               // 是否启用
	Description  string `json:"description"`                           // 描述
}

// CreateRoleResp represents role creation response.
type CreateRoleResp struct {
	ID uint `json:"id"`
}

// ListRolesReq represents role list query parameters.
type ListRolesReq struct {
	PageReq
	SortReq
	Key     string `form:"key" binding:"omitempty,max=100"`
	Enabled *bool  `form:"enabled"`
}

// ListRolesResp represents role list with pagination info.
type ListRolesResp struct {
	Results    []ListRolesRespResult `json:"results"`
	Pagination PageResp              `json:"pagination"`
}

// ListRolesRespResult represents a single role in the list.
type ListRolesRespResult struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`          // 角色名称
	Slug         string `json:"slug"`          // 角色标识
	Code         string `json:"code"`          // 角色编号
	DisplayOrder int32  `json:"display_order"` // 显示顺序
	Enabled      bool   `json:"enabled"`       // 是否启用
	Description  string `json:"description"`   // 描述
	UserCount    int64  `json:"user_count"`    // 用户数量
}

// GetRoleReq represents a request to get a specific role.
type GetRoleReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// GetRoleResp represents a role object in responses.
type GetRoleResp struct {
	ID           uint   `json:"id"`            // ID
	CreatedAt    string `json:"created_at"`    // 创建时间
	Name         string `json:"name"`          // 角色名称
	Slug         string `json:"slug"`          // 角色标识
	Code         string `json:"code"`          // 角色编号
	DisplayOrder int32  `json:"display_order"` // 显示顺序
	Description  string `json:"description"`   // 描述
	Enabled      bool   `json:"enabled"`       // 是否启用
	UserCount    int64  `json:"user_count"`    // 用户数量
}

type GetRoleFormDataReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// GetRoleFormDataResp represents data needed for role forms (create/update).
type GetRoleFormDataResp struct {
	ID           uint   `json:"id"`            // ID
	Code         string `json:"code"`          // 角色编号
	Slug         string `json:"slug"`          // 角色标识
	Name         string `json:"name"`          // 角色名称
	DisplayOrder int32  `json:"display_order"` // 显示顺序
	Enabled      bool   `json:"enabled"`       // 是否启用
	Description  string `json:"description"`   // 描述
}

// UpdateRoleReq represents role update data.
type UpdateRoleReq struct {
	ID           uint   `uri:"id" binding:"required" swaggerignore:"true"`
	Code         string `json:"code" binding:"required,min=1,max=100"` // 角色编号
	Name         string `json:"name" binding:"required,min=1,max=100"` // 角色名称
	DisplayOrder int32  `json:"display_order" binding:"required"`      // 显示顺序
	Enabled      bool   `json:"enabled"`                               // 是否启用
	Description  string `json:"description"`                           // 描述
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
	SortReq
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

// BatchDeleteRolesReq represents a request to batch delete roles.
type BatchDeleteRolesReq struct {
	IDs []uint `json:"ids" binding:"required,min=1,dive,required"`
}

// BatchDeleteRolesResp represents batch delete roles response.
type BatchDeleteRolesResp struct {
	SuccessCount int      `json:"success_count"` // 成功删除数量
	FailedCount  int      `json:"failed_count"`  // 失败数量
	FailedItems  []string `json:"failed_items"`  // 失败的信息（格式：名称: 错误原因）
}
