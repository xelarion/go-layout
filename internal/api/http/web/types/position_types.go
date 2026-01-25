package types

// Common position-related types

// CreatePositionReq represents position creation request.
type CreatePositionReq struct {
	Code           string `json:"code" binding:"required,min=1,max=100"`    // 角色编号
	Name           string `json:"name" binding:"required,min=1,max=100"`    // 角色名称
	OrganizationID uint   `json:"organization_id" binding:"required,min=1"` // 组织ID
	Remark         string `json:"remark" binding:"omitempty,max=1000"`      // 备注
}

// CreatePositionResp represents position creation response.
type CreatePositionResp struct {
	ID uint `json:"id"`
}

// ListPositionsReq represents position list query parameters.
type ListPositionsReq struct {
	PageReq
	SortReq
	Key            string `form:"key" binding:"omitempty,max=100"`           // 搜索关键词
	OrganizationID uint   `form:"organization_id" binding:"omitempty,min=1"` // 所属组织
}

// ListPositionsResp represents position list with pagination info.
type ListPositionsResp struct {
	Results    []ListPositionsRespResult `json:"results"`
	Pagination PageResp                  `json:"pagination"`
}

// ListPositionsRespResult represents a single position in the list.
type ListPositionsRespResult struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`              // 角色名称
	Code             string `json:"code"`              // 角色编号
	OrganizationName string `json:"organization_name"` // 组织名称
	UserCount        int64  `json:"user_count"`        // 用户数量
	Remark           string `json:"remark"`            // 备注
}

type GetPositionFormDataReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// GetPositionFormDataResp represents data needed for position forms (create/update).
type GetPositionFormDataResp struct {
	ID             uint   `json:"id"`
	Code           string `json:"code"`            // 角色编号
	Name           string `json:"name"`            // 角色名称
	OrganizationID uint   `json:"organization_id"` // 组织ID
	Remark         string `json:"remark"`          // 备注
}

// UpdatePositionReq represents position update data.
type UpdatePositionReq struct {
	ID             uint   `uri:"id" binding:"required" swaggerignore:"true"`
	Code           string `json:"code" binding:"required,min=1,max=100"` // 角色编号
	Name           string `json:"name" binding:"required,min=1,max=100"` // 角色名称
	OrganizationID uint   `json:"organization_id" binding:"required"`    // 组织ID
	Remark         string `json:"remark" binding:"omitempty,max=1000"`   // 备注
}

// UpdatePositionResp represents position update response.
type UpdatePositionResp struct {
}

// DeletePositionReq represents a request to delete a position.
type DeletePositionReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// DeletePositionResp represents position delete response.
type DeletePositionResp struct {
}

// GetPositionOptionsReq represents a request to get position options.
type GetPositionOptionsReq struct {
	SortReq
}

// GetPositionOptionsResp represents position options response.
type GetPositionOptionsResp struct {
	Results []GetPositionOptionsRespResult `json:"results"`
}

type GetPositionOptionsRespResult struct {
	GetOptionsRespResult
	OrganizationName string `json:"organization_name"` // 组织名称
}

// BatchDeletePositionsReq represents a request to batch delete positions.
type BatchDeletePositionsReq struct {
	IDs []uint `json:"ids" binding:"required,min=1,dive,required"`
}

// BatchDeletePositionsResp represents batch delete positions response.
type BatchDeletePositionsResp struct {
	SuccessCount int      `json:"success_count"` // 成功删除数量
	FailedCount  int      `json:"failed_count"`  // 失败数量
	FailedItems  []string `json:"failed_items"`  // 失败的信息（格式：名称: 错误原因）
}
