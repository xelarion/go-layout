package model

import (
	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model/gen"
)

// User represents a user model.
type User struct {
	gen.User

	Gender enum.UserGender `gorm:"column:gender;type:smallint;not null;comment:性别" json:"gender"` // 性别
}

// GetResourceType 实现 Resource 接口
func (u *User) GetResourceType() string {
	return "user"
}

// GetResourceID 实现 Resource 接口
func (u *User) GetResourceID() uint {
	return u.ID
}

// GetCreatedBy 实现 DataScopeRecord 接口
func (u *User) GetCreatedBy() uint {
	return u.CreatedBy
}

// GetCreatedRoleID 实现 DataScopeRecord 接口
func (u *User) GetCreatedRoleID() uint {
	return u.CreatedRoleID
}

// GetAssignedTo 实现 DataScopeRecord 接口
func (u *User) GetAssignedTo() *uint {
	return nil // user 表没有 assigned_to 字段
}

// GetDataScopeFieldMap 实现 DataScopeRecord 接口
// 返回 users 表支持的数据权限字段映射
func (u *User) GetDataScopeFieldMap() map[string]string {
	return map[string]string{
		"created_by":      "created_by",      // 创建人字段
		"created_role_id": "created_role_id", // 创建角色字段
		// users 表不支持 assigned_to，所以不在此定义
	}
}
