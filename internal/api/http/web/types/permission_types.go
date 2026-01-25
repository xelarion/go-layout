package types

// GetPermissionsTreeReq is the request for getting the permission tree
type GetPermissionsTreeReq struct{}

// GetPermissionsTreeResp is the response for getting the permission tree
type GetPermissionsTreeResp struct {
	Permissions []*TreeNodeResp `json:"permissions"`
}

// GetRolePermissionsReq is the request for getting role permissions
type GetRolePermissionsReq struct {
	RoleID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// GetRolePermissionsResp is the response for getting role permissions
type GetRolePermissionsResp struct {
	Permissions []string                          `json:"permissions"`
	DataScopes  []GetRolePermissionsRespDataScope `json:"data_scopes"`
}

type GetRolePermissionsRespDataScope struct {
	Permissions []string `json:"permissions"`
	Type        string   `json:"type"`
}

// UpdateRolePermissionsReq is the request for updating role permissions
type UpdateRolePermissionsReq struct {
	RoleID      uint                                `uri:"id" binding:"required" swaggerignore:"true"`
	Permissions []string                            `json:"permissions" binding:"required"`
	DataScopes  []UpdateRolePermissionsReqDataScope `json:"data_scopes" binding:"required"`
}

type UpdateRolePermissionsReqDataScope struct {
	Permissions []string `json:"permissions" binding:"required"`
	Type        string   `json:"type" binding:"required"`
}

// UpdateRolePermissionsResp is the response for updating role permissions
type UpdateRolePermissionsResp struct{}
