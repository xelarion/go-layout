package types

// ExportResp 导出响应（统一返回任务信息）
//
// 使用方式：
// 1. 调用导出接口（如 POST /users/export），获取任务ID和IsAsync标识
// 2. 根据IsAsync判断：
//   - IsAsync=false（同步）：DownloadURL 已包含在响应中，直接使用该URL下载文件
//   - IsAsync=true（异步）：轮询任务列表查询状态，完成后调用下载URL接口（GET /export_tasks/:id/download_url）获取预签名URL
//
// 3. 任务和文件会在 ExpireDays 后自动清理，也可手动删除
type ExportResp struct {
	ID          uint   `json:"id"`                     // 任务ID
	IsAsync     bool   `json:"is_async"`               // 是否异步导出
	TotalCount  int    `json:"total_count"`            // 总记录数
	DownloadURL string `json:"download_url,omitempty"` // 下载URL（仅同步导出时返回）
}

// ListExportTasksReq 导出任务列表请求
type ListExportTasksReq struct {
	PageReq
	SortReq
	TaskType string `form:"task_type" binding:"omitempty"` // 任务类型筛选
	Status   string `form:"status" binding:"omitempty"`    // 状态筛选
}

// ListExportTasksResp 导出任务列表响应
type ListExportTasksResp struct {
	Results    []ExportTaskResult `json:"results"`
	Pagination PageResp           `json:"pagination"`
}

// ExportTaskResult 导出任务结果
type ExportTaskResult struct {
	ID           uint   `json:"id"`            // 任务ID
	TaskType     string `json:"task_type"`     // 任务类型
	Filename     string `json:"filename"`      // 文件名
	Status       string `json:"status"`        // 状态
	StatusSlug   string `json:"status_slug"`   // 状态 slug
	IsAsync      bool   `json:"is_async"`      // 是否异步导出
	TotalCount   int    `json:"total_count"`   // 总记录数
	FileSize     int64  `json:"file_size"`     // 文件大小
	ErrorMessage string `json:"error_message"` // 错误信息
	CreatedAt    string `json:"created_at"`    // 创建时间
	StartedAt    string `json:"started_at"`    // 开始处理时间
	CompletedAt  string `json:"completed_at"`  // 完成时间
}

// GetExportFileDownloadURLReq 获取导出文件下载URL请求
type GetExportFileDownloadURLReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// GetExportFileDownloadURLResp 获取导出文件下载URL响应
type GetExportFileDownloadURLResp struct {
	URL string `json:"url"` // 预签名URL（文件名已编码在URL中）
}

// DeleteExportTaskReq 删除导出任务请求
type DeleteExportTaskReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// DeleteExportTaskResp 删除导出任务响应
type DeleteExportTaskResp struct{}
