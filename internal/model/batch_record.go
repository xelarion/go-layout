package model

import (
	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model/gen"
)

type BatchRecord struct {
	gen.BatchRecord

	OperationType enum.BatchRecordOperationType `gorm:"column:operation_type;type:smallint;not null;comment:操作类型" json:"operation_type"` // 操作类型
	Status        enum.BatchRecordStatus        `gorm:"column:status;type:smallint;not null;comment:状态" json:"status"`                   // 状态
}

// BatchRecordDataNone 无数据(示例代码，其他类型请自定义)
type BatchRecordDataNone struct {
}

func (b *BatchRecord) GetResourceType() string {
	return "batch_record"
}

func (b *BatchRecord) GetResourceID() uint {
	return b.ID
}

// GetCreatedBy 实现 DataScopeRecord 接口
func (b *BatchRecord) GetCreatedBy() uint {
	return b.CreatedBy
}

// GetCreatedRoleID 实现 DataScopeRecord 接口
func (b *BatchRecord) GetCreatedRoleID() uint {
	return b.CreatedRoleID
}

// GetAssignedTo 实现 DataScopeRecord 接口
func (b *BatchRecord) GetAssignedTo() *uint {
	return nil // batch_record 表没有 assigned_to 字段
}

// GetDataScopeFieldMap 实现 DataScopeRecord 接口
// 返回 batch_records 表支持的数据权限字段映射
func (b *BatchRecord) GetDataScopeFieldMap() map[string]string {
	return map[string]string{
		"created_by":      "created_by",      // 创建人字段
		"created_role_id": "created_role_id", // 创建角色字段
		// batch_records 表不支持 assigned_to，所以不在此定义
	}
}

func (b *BatchRecord) IsStatusPending() bool {
	return b.Status == enum.BatchRecordStatusPending
}

func (b *BatchRecord) IsStatusProcessing() bool {
	return b.Status == enum.BatchRecordStatusProcessing
}

func (b *BatchRecord) IsStatusFinished() bool {
	return b.Status == enum.BatchRecordStatusSuccess || b.Status == enum.BatchRecordStatusFailed
}
