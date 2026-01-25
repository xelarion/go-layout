package model

import (
	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model/gen"
)

type BatchDetail struct {
	gen.BatchDetail

	Status enum.BatchDetailStatus `gorm:"column:status;type:smallint;not null;comment:状态" json:"status"` // 状态
}

// BatchDetailDataNone 无数据(示例代码，其他类型请自定义)
type BatchDetailDataNone struct {
}

func (b *BatchDetail) GetResourceType() string {
	return "batch_detail"
}

func (b *BatchDetail) GetResourceID() uint {
	return b.ID
}

func (b *BatchDetail) IsStatusPending() bool {
	return b.Status == enum.BatchDetailStatusPending
}

func (b *BatchDetail) IsStatusProcessing() bool {
	return b.Status == enum.BatchDetailStatusProcessing
}

func (b *BatchDetail) IsStatusFinished() bool {
	return b.Status == enum.BatchDetailStatusSuccess || b.Status == enum.BatchDetailStatusFailed
}
