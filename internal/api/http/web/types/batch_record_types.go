package types

type ListBatchRecordsReq struct {
	PageReq
	SortReq
	CreatedBy      uint   `form:"created_by" binding:"omitempty,min=1"` // 创建人
	OperationType  string `form:"operation_type" binding:"omitempty"`   // 操作类型
	Status         string `form:"status" binding:"omitempty"`           // 状态
	CreatedAtStart string `form:"created_at_start" binding:"omitempty"` // 创建时间开始
	CreatedAtEnd   string `form:"created_at_end" binding:"omitempty"`   // 创建时间结束
}

type ListBatchRecordsResp struct {
	Results    []ListBatchRecordsRespResult `json:"results"`
	Pagination PageResp                     `json:"pagination"`
}

type ListBatchRecordsRespResult struct {
	ID                uint   `json:"id"`                  // ID
	CreatedAt         string `json:"created_at"`          // 创建时间
	CreatedBy         string `json:"created_by"`          // 创建人
	OperationType     string `json:"operation_type"`      // 操作类型
	OperationTypeSlug string `json:"operation_type_slug"` // 操作类型 slug
	Status            string `json:"status"`              // 状态
	StatusSlug        string `json:"status_slug"`         // 状态 slug
	FailureReason     string `json:"failure_reason"`      // 失败原因
}

type GetBatchRecordOperationTypeOptionsReq struct {
}

type GetBatchRecordOperationTypeOptionsResp struct {
	Results []GetBatchRecordOperationTypeOptionsRespResult `json:"results"`
}

type GetBatchRecordOperationTypeOptionsRespResult struct {
	GetOptionsRespResult
}

type GetBatchRecordReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

type GetBatchRecordResp struct {
	ID                uint           `json:"id"`                  // ID
	CreatedAt         string         `json:"created_at"`          // 创建时间
	CreatedBy         string         `json:"created_by"`          // 创建人
	OperationType     string         `json:"operation_type"`      // 操作类型
	OperationTypeSlug string         `json:"operation_type_slug"` // 操作类型 slug
	Status            string         `json:"status"`              // 状态
	StatusSlug        string         `json:"status_slug"`         // 状态 slug
	FailureReason     string         `json:"failure_reason"`      // 失败原因
	ExtraData         map[string]any `json:"extra_data"`          // 额外数据(如非通用字段)
}
