package permission

import "github.com/xelarion/go-layout/internal/enum"

// RoleDataScope represents a role permissions data scope
// 角色数据权限范围配置
type RoleDataScope struct {
	Permissions []string               `json:"permissions"` // 受此数据范围限制的权限列表
	Type        enum.RoleDataScopeType `json:"type"`        // 数据范围类型：all, same_role, self
}

// RoleWithDataScopes 角色数据权限接口
// 任何角色类型都应实现此接口，以便进行数据权限检查
type RoleWithDataScopes interface {
	IsSuperAdmin() bool
	GetDataScopes() []RoleDataScope
}

// DataScopeRecord 数据权限记录接口
// 任何需要数据权限检查的结构体都应实现此接口
type DataScopeRecord interface {
	GetCreatedBy() uint
	GetCreatedRoleID() uint
	GetAssignedTo() *uint // 返回指针，因为可能为 nil

	// GetDataScopeFieldMap 返回数据权限支持的字段映射
	// key: 逻辑字段名（DataScopeFilter 使用）
	// value: 数据库字段名（SQL 查询使用）
	// 如果表不支持某个字段（如 assigned_to），则不要在 map 中返回
	GetDataScopeFieldMap() map[string]string
}
