package permission

import (
	"context"

	appctx "github.com/xelarion/go-layout/internal/context"
	"github.com/xelarion/go-layout/internal/enum"
)

// DataScopeFilter 数据权限过滤器
type DataScopeFilter struct {
	role    RoleWithDataScopes
	current *appctx.Current
	record  DataScopeRecord // 用于获取字段映射
}

// NewDataScopeFilter 创建数据权限过滤器
func NewDataScopeFilter(ctx context.Context, role RoleWithDataScopes, record DataScopeRecord) *DataScopeFilter {
	return &DataScopeFilter{
		role:    role,
		current: appctx.GetCurrent(ctx),
		record:  record,
	}
}

// Apply 应用数据权限到查询过滤条件
// 自动从 context 获取权限列表，处理所有权限的数据范围（取并集）
func (f *DataScopeFilter) Apply(ctx context.Context, filters map[string]any) bool {
	// 从 context 获取当前请求的权限列表
	permissions := f.current.ReqPermissions
	if len(permissions) == 0 {
		return false // 没有权限信息
	}

	// 超级管理员不过滤
	if f.role.IsSuperAdmin() {
		return true
	}

	// 收集所有匹配的数据权限配置（对所有请求权限）
	var allMatchedScopes []RoleDataScope
	for _, reqPerm := range permissions {
		for _, scope := range f.role.GetDataScopes() {
			if f.hasPermission(scope.Permissions, reqPerm) {
				allMatchedScopes = append(allMatchedScopes, scope)
			}
		}
	}

	if len(allMatchedScopes) == 0 {
		return false
	}

	return f.applyScopes(allMatchedScopes, filters)
}

// hasPermission 检查权限列表中是否包含目标权限
func (f *DataScopeFilter) hasPermission(permissions []string, target string) bool {
	for _, p := range permissions {
		if p == target || p == "*" {
			return true
		}
	}
	return false
}

// applyScopes 应用多个数据权限范围（OR 逻辑，取并集）
func (f *DataScopeFilter) applyScopes(scopes []RoleDataScope, filters map[string]any) bool {
	// 获取字段映射
	fieldMap := f.record.GetDataScopeFieldMap()

	// 收集过滤条件
	hasSameRole := false
	hasSelf := false

	for _, scope := range scopes {
		switch scope.Type {
		case enum.RoleDataScopeTypeAll: // 如果任意一个 scope 是 "all"，则不添加任何过滤
			return true
		case enum.RoleDataScopeTypeSameRole:
			hasSameRole = true
		case enum.RoleDataScopeTypeSelf:
			hasSelf = true
		}
	}

	// 应用过滤条件（OR 逻辑）
	if hasSameRole && hasSelf {
		// 同一角色 OR 本人
		var conditions []map[string]any

		// 添加 created_role_id 条件
		if dbField, ok := fieldMap["created_role_id"]; ok {
			conditions = append(conditions, map[string]any{
				dbField: f.current.RoleID,
			})
		}

		// 添加 created_by 条件
		if dbField, ok := fieldMap["created_by"]; ok {
			conditions = append(conditions, map[string]any{
				dbField: f.current.UserID,
			})
		}

		// 添加 assigned_to 条件（如果表支持）
		if dbField, ok := fieldMap["assigned_to"]; ok {
			conditions = append(conditions, map[string]any{
				dbField: f.current.UserID,
			})
		}

		if len(conditions) > 0 {
			filters["_data_scope_or"] = conditions
		}
	} else if hasSameRole {
		// 仅同一角色
		if dbField, ok := fieldMap["created_role_id"]; ok {
			filters[dbField] = f.current.RoleID
		}
	} else if hasSelf {
		// 仅本人（创建人 OR 被指派人）
		var conditions []map[string]any

		// 添加 created_by 条件
		if dbField, ok := fieldMap["created_by"]; ok {
			conditions = append(conditions, map[string]any{
				dbField: f.current.UserID,
			})
		}

		// 添加 assigned_to 条件（如果表支持）
		if dbField, ok := fieldMap["assigned_to"]; ok {
			conditions = append(conditions, map[string]any{
				dbField: f.current.UserID,
			})
		}

		if len(conditions) > 0 {
			filters["_data_scope_or"] = conditions
		}
	}

	return true
}

// CheckRecordAccess 检查是否有权访问单条记录
// 使用接口类型，类型安全
func (f *DataScopeFilter) CheckRecordAccess(ctx context.Context, record DataScopeRecord) bool {
	// 从 context 获取当前请求的权限列表
	permissions := f.current.ReqPermissions
	if len(permissions) == 0 {
		return false
	}

	// 超级管理员不限制
	if f.role.IsSuperAdmin() {
		return true
	}

	// 收集所有匹配的数据权限配置
	var allMatchedScopes []RoleDataScope
	for _, reqPerm := range permissions {
		for _, scope := range f.role.GetDataScopes() {
			if f.hasPermission(scope.Permissions, reqPerm) {
				allMatchedScopes = append(allMatchedScopes, scope)
			}
		}
	}

	if len(allMatchedScopes) == 0 {
		return false
	}

	// 检查记录是否符合任意一个 scope（OR 逻辑）
	for _, scope := range allMatchedScopes {
		if f.matchScope(scope, record) {
			return true
		}
	}

	return false
}

// matchScope 检查记录是否匹配指定的数据权限范围
func (f *DataScopeFilter) matchScope(scope RoleDataScope, record DataScopeRecord) bool {
	switch scope.Type {
	case enum.RoleDataScopeTypeAll:
		return true

	case enum.RoleDataScopeTypeSameRole:
		return record.GetCreatedRoleID() == f.current.RoleID

	case enum.RoleDataScopeTypeSelf:
		// 检查是否是创建人
		if record.GetCreatedBy() == f.current.UserID {
			return true
		}
		// 检查是否是被指派人
		if assignedTo := record.GetAssignedTo(); assignedTo != nil {
			if *assignedTo == f.current.UserID {
				return true
			}
		}
	}
	return false
}
