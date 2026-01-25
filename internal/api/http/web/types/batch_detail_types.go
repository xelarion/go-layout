package types

type ListBatchDetailsReq struct {
	PageReq
	SortReq
	BatchRecordID uint   `form:"batch_record_id" binding:"required,min=1"` // 批次记录ID
	Status        string `form:"status" binding:"omitempty"`               // 状态
	Key           string `form:"key" binding:"omitempty"`                  // 关键字段
}

type ListBatchDetailsResp struct {
	Results    []ListBatchDetailsRespResult `json:"results"`
	Pagination PageResp                     `json:"pagination"`
}

type ListBatchDetailsRespResult struct {
	ID            uint           `json:"id"`             // ID
	CreatedAt     string         `json:"created_at"`     // 创建时间
	Key           string         `json:"key"`            // 关键字段(用于标识此行数据或搜索)
	Status        string         `json:"status"`         // 状态
	StatusSlug    string         `json:"status_slug"`    // 状态 slug
	FailureReason string         `json:"failure_reason"` // 失败原因
	ExtraData     map[string]any `json:"extra_data"`     // 额外数据(如非通用字段)
}
