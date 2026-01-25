package enum

// Role types
const (
	RoleSuperAdmin = "super_admin"
)

// RoleDataScopeType 数据权限类型
type RoleDataScopeType string

const (
	RoleDataScopeTypeAll      RoleDataScopeType = "all"       // 全部
	RoleDataScopeTypeSameRole RoleDataScopeType = "same_role" // 同一角色
	RoleDataScopeTypeSelf     RoleDataScopeType = "self"      // 仅本人，包括创建人或被指派人
)
