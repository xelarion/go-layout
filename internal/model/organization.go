package model

import (
	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model/gen"
)

// Organization represents a Organization
type Organization struct {
	gen.Organization

	Type enum.OrganizationType `gorm:"column:type;type:smallint;not null;comment:组织类型" json:"type"` // 组织类型
}

// GetResourceType 实现 Resource 接口
func (r *Organization) GetResourceType() string {
	return "organization"
}

// GetResourceID 实现 Resource 接口
func (r *Organization) GetResourceID() uint {
	return r.ID
}

// GetCreatedBy 实现 DataScopeRecord 接口
func (r *Organization) GetCreatedBy() uint {
	return r.CreatedBy
}

// GetCreatedRoleID 实现 DataScopeRecord 接口
func (r *Organization) GetCreatedRoleID() uint {
	return r.CreatedRoleID
}

// GetAssignedTo 实现 DataScopeRecord 接口
func (r *Organization) GetAssignedTo() *uint {
	return nil // organizations 表没有 assigned_to 字段
}

// GetDataScopeFieldMap 实现 DataScopeRecord 接口
// 返回 organizations 表支持的数据权限字段映射
func (r *Organization) GetDataScopeFieldMap() map[string]string {
	return map[string]string{
		"created_by":      "created_by",      // 创建人字段
		"created_role_id": "created_role_id", // 创建角色字段
		// organizations 表不支持 assigned_to，所以不在此定义
	}
}
