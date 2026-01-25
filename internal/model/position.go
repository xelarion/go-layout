package model

import (
	"github.com/xelarion/go-layout/internal/model/gen"
)

// Position represents a Position
type Position struct {
	gen.Position
}

// GetResourceType 实现 Resource 接口
func (r *Position) GetResourceType() string {
	return "position"
}

// GetResourceID 实现 Resource 接口
func (r *Position) GetResourceID() uint {
	return r.ID
}

// GetCreatedBy 实现 DataScopeRecord 接口
func (r *Position) GetCreatedBy() uint {
	return r.CreatedBy
}

// GetCreatedRoleID 实现 DataScopeRecord 接口
func (r *Position) GetCreatedRoleID() uint {
	return r.CreatedRoleID
}

// GetAssignedTo 实现 DataScopeRecord 接口
func (r *Position) GetAssignedTo() *uint {
	return nil // positions 表没有 assigned_to 字段
}

// GetDataScopeFieldMap 实现 DataScopeRecord 接口
// 返回 positions 表支持的数据权限字段映射
func (r *Position) GetDataScopeFieldMap() map[string]string {
	return map[string]string{
		"created_by":      "created_by",      // 创建人字段
		"created_role_id": "created_role_id", // 创建角色字段
		// positions 表不支持 assigned_to，所以不在此定义
	}
}
