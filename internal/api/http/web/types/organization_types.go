package types

// Common organization-related types

// CreateOrganizationReq represents organization creation request.
type CreateOrganizationReq struct {
	Code     string `json:"code" binding:"required,min=1,max=100"` // 角色编号
	Name     string `json:"name" binding:"required,min=1,max=100"` // 角色名称
	ParentID *uint  `json:"parent_id" binding:"omitempty,min=1"`   // 父级ID
	Type     string `json:"type" binding:"required"`               // 组织类型
	OwnerID  uint   `json:"owner_id" binding:"required,min=1"`     // 负责人ID
	Remark   string `json:"remark" binding:"omitempty,max=1000"`   // 备注
}

// CreateOrganizationResp represents organization creation response.
type CreateOrganizationResp struct {
	ID uint `json:"id"`
}

// ListOrganizationsReq represents organization list query parameters.
type ListOrganizationsReq struct {
	SortReq
	Key     string `form:"key" binding:"omitempty,max=100"` // 搜索关键词
	Type    string `form:"type" binding:"omitempty"`        // 组织类型
	OwnerID uint   `form:"owner_id" binding:"omitempty"`    // 负责人ID
}

// ListOrganizationsResp represents organization list with pagination info.
type ListOrganizationsResp struct {
	Results []*ListOrganizationsRespResult `json:"results"`
}

// ListOrganizationsRespResult represents a single organization in the list.
type ListOrganizationsRespResult struct {
	ID        uint                           `json:"id"`
	Name      string                         `json:"name"`       // 角色名称
	Code      string                         `json:"code"`       // 角色编号
	ParentID  *uint                          `json:"parent_id"`  // 父级ID
	Type      string                         `json:"type"`       // 组织类型
	Owner     string                         `json:"owner"`      // 负责人
	UserCount int64                          `json:"user_count"` // 员工人数
	Remark    string                         `json:"remark"`     // 备注
	Children  []*ListOrganizationsRespResult `json:"children"`   // 子级
}

type GetOrganizationFormDataReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// GetOrganizationFormDataResp represents data needed for organization forms (create/update).
type GetOrganizationFormDataResp struct {
	ID       uint   `json:"id"`
	Code     string `json:"code"`      // 角色编号
	Name     string `json:"name"`      // 角色名称
	ParentID *uint  `json:"parent_id"` // 父级ID
	Type     string `json:"type"`      // 组织类型
	OwnerID  uint   `json:"owner_id"`  // 负责人ID
	Remark   string `json:"remark"`    // 备注
}

// UpdateOrganizationReq represents organization update data.
type UpdateOrganizationReq struct {
	ID       uint   `uri:"id" binding:"required" swaggerignore:"true"`
	Code     string `json:"code" binding:"required,min=1,max=100"` // 角色编号
	Name     string `json:"name" binding:"required,min=1,max=100"` // 角色名称
	ParentID *uint  `json:"parent_id" binding:"omitempty,min=1"`   // 父级ID
	Type     string `json:"type" binding:"required"`               // 组织类型
	OwnerID  uint   `json:"owner_id" binding:"required"`           // 负责人ID
	Remark   string `json:"remark" binding:"omitempty,max=1000"`   // 备注
}

// UpdateOrganizationResp represents organization update response.
type UpdateOrganizationResp struct {
}

// DeleteOrganizationReq represents a request to delete a organization.
type DeleteOrganizationReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// DeleteOrganizationResp represents organization delete response.
type DeleteOrganizationResp struct {
}

// GetOrganizationsTreeReq is the request for getting the organization tree
type GetOrganizationsTreeReq struct{}

// GetOrganizationsTreeResp is the response for getting the organization tree
type GetOrganizationsTreeResp struct {
	Results []*TreeNodeResp `json:"results"`
}
