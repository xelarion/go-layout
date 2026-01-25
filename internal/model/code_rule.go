package model

import (
	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model/gen"
)

// CodeRule represents a code_rule model.
type CodeRule struct {
	gen.CodeRule

	TimeRule     enum.CodeRuleTimeRule     `gorm:"column:time_rule;type:smallint;not null;comment:时间规则" json:"time_rule"`         // 时间规则
	GenerateType enum.CodeRuleGenerateType `gorm:"column:generate_type;type:smallint;not null;comment:生成类型" json:"generate_type"` // 生成类型
}

// GetResourceType 实现 Resource 接口
func (u *CodeRule) GetResourceType() string {
	return "code_rule"
}

// GetResourceID 实现 Resource 接口
func (u *CodeRule) GetResourceID() uint {
	return u.ID
}

// GetCreatedBy 实现 DataScopeRecord 接口
func (u *CodeRule) GetCreatedBy() uint {
	return u.CreatedBy
}

// GetCreatedRoleID 实现 DataScopeRecord 接口
func (u *CodeRule) GetCreatedRoleID() uint {
	return u.CreatedRoleID
}

// GetAssignedTo 实现 DataScopeRecord 接口
func (u *CodeRule) GetAssignedTo() *uint {
	return nil // code_rule 表没有 assigned_to 字段
}

// GetDataScopeFieldMap 实现 DataScopeRecord 接口
// 返回 code_rules 表支持的数据权限字段映射
func (u *CodeRule) GetDataScopeFieldMap() map[string]string {
	return map[string]string{
		"created_by":      "created_by",      // 创建人字段
		"created_role_id": "created_role_id", // 创建角色字段
		// code_rules 表不支持 assigned_to，所以不在此定义
	}
}
