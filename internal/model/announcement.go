package model

import (
	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model/gen"
)

type Announcement struct {
	gen.Announcement

	Status enum.AnnouncementStatus `gorm:"column:status;type:smallint;not null;comment:状态" json:"status"` // 状态
}

// GetResourceType 实现 Resource 接口
func (e *Announcement) GetResourceType() string {
	return "announcement"
}

// GetResourceID 实现 Resource 接口
func (e *Announcement) GetResourceID() uint {
	return e.ID
}

// GetCreatedBy 实现 DataScopeRecord 接口
func (e *Announcement) GetCreatedBy() uint {
	return e.CreatedBy
}

// GetCreatedRoleID 实现 DataScopeRecord 接口
func (e *Announcement) GetCreatedRoleID() uint {
	return e.CreatedRoleID
}

// GetAssignedTo 实现 DataScopeRecord 接口
func (e *Announcement) GetAssignedTo() *uint {
	return nil // announcement 表没有 assigned_to 字段
}

// GetDataScopeFieldMap 实现 DataScopeRecord 接口
// 返回 announcement 表支持的数据权限字段映射
func (e *Announcement) GetDataScopeFieldMap() map[string]string {
	return map[string]string{
		"created_by":      "created_by",      // 创建人字段
		"created_role_id": "created_role_id", // 创建角色字段
		// announcement 表不支持 assigned_to，所以不在此定义
	}
}

// IsPublished 是否已发布
func (e *Announcement) IsPublished() bool {
	return e.Status == enum.AnnouncementStatusPublished
}
