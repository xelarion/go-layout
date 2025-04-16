package types

// GetPermissionTreeReq is the request for getting the permission tree
type GetPermissionTreeReq struct{}

// GetPermissionTreeResp is the response for getting the permission tree
type GetPermissionTreeResp struct {
	Permissions []*GetPermissionTreeRespPermission `json:"permissions"`
}

// GetPermissionTreeRespPermission represents a permission node in the response
type GetPermissionTreeRespPermission struct {
	ID       string                             `json:"id"`                 // Permission identifier
	Name     string                             `json:"name"`               // Permission name
	Children []*GetPermissionTreeRespPermission `json:"children,omitempty"` // Child permissions
}

// GetRolePermissionsReq is the request for getting role permissions
type GetRolePermissionsReq struct {
	RoleID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// GetRolePermissionsResp is the response for getting role permissions
type GetRolePermissionsResp struct {
	Permissions []string `json:"permissions"`
}

// UpdateRolePermissionsReq is the request for updating role permissions
type UpdateRolePermissionsReq struct {
	RoleID      uint     `uri:"id" binding:"required" swaggerignore:"true"`
	Permissions []string `json:"permissions" binding:"required"`
}

// UpdateRolePermissionsResp is the response for updating role permissions
type UpdateRolePermissionsResp struct{}
