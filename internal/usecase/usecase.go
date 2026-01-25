package usecase

import (
	"context"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Transaction interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
	IsInTransaction(ctx context.Context) bool
}

// CreateRecordCommonParams 创建记录通用字段
type CreateRecordCommonParams struct {
	CreatedBy     uint // Creator user ID
	CreatedRoleID uint // Creator user role ID
	AssignedTo    uint // Assigned user ID
}

type UpdateRecordCommonParams struct {
	UpdatedBy  uint // Updated user ID
	AssignedTo uint // Assigned user ID
}

// BatchItemCommonResult 批量处理(如 批量删除)的结果
type BatchItemCommonResult struct {
	ID         uint
	Identifier string // 标识
	Success    bool
	Error      error
}
