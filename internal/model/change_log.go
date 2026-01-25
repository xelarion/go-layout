package model

import (
	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model/gen"
)

type ChangeLog struct {
	gen.ChangeLog

	Action enum.ChangeLogAction `gorm:"column:action;type:smallint;not null;comment:操作类型" json:"action"` // 操作类型

	Operator *User `gorm:"foreignKey:OperatorID;references:ID;->" json:"operator"` // 操作人
}

// GetResourceType 实现 Resource 接口
func (e *ChangeLog) GetResourceType() string {
	return "change_log"
}

// GetResourceID 实现 Resource 接口
func (e *ChangeLog) GetResourceID() uint {
	return e.ID
}
