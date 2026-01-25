package types

// Common announcement-related types

// CreateAnnouncementReq represents announcement creation request.
type CreateAnnouncementReq struct {
	Title   string `json:"title" binding:"required,min=1,max=100"` // 标题名称
	Content string `json:"content" binding:"required"`             // 内容
	Status  string `json:"status" binding:"required"`              // 状态
}

// CreateAnnouncementResp represents announcement creation response.
type CreateAnnouncementResp struct {
	ID uint `json:"id"`
}

// ListAnnouncementsReq represents announcement list query parameters.
type ListAnnouncementsReq struct {
	PageReq
	SortReq
	Key              string `form:"key" binding:"omitempty,max=100"`        // 搜索关键字
	PublisherID      uint   `form:"publisher_id" binding:"omitempty"`       // 发布人ID
	PublishedAtStart string `form:"published_at_start" binding:"omitempty"` // 发布时间开始
	PublishedAtEnd   string `form:"published_at_end" binding:"omitempty"`   // 发布时间结束
	Status           string `form:"status" binding:"omitempty"`             // 状态
}

// ListAnnouncementsResp represents announcement tree list without pagination.
type ListAnnouncementsResp struct {
	Results    []*ListAnnouncementsRespResult `json:"results"`
	Pagination PageResp                       `json:"pagination"`
}

// ListAnnouncementsRespResult represents a single announcement in the list.
type ListAnnouncementsRespResult struct {
	ID          uint   `json:"id"`           // ID
	Title       string `json:"title"`        // 标题名称
	Status      string `json:"status"`       // 状态
	StatusSlug  string `json:"status_slug"`  // 状态标识
	Publisher   string `json:"publisher"`    // 发布人
	PublishedAt string `json:"published_at"` // 发布时间
}

// GetAnnouncementReq represents a request to get a specific announcement.
type GetAnnouncementReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"` // ID
}

// GetAnnouncementResp represents a announcement object in responses.
type GetAnnouncementResp struct {
	ID          uint   `json:"id"`           // ID
	Title       string `json:"title"`        // 标题名称
	Content     string `json:"content"`      // 内容
	PublishedAt string `json:"published_at"` // 发布时间
}

type GetAnnouncementFormDataReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"` // ID
}

// GetAnnouncementFormDataResp represents data needed for announcement forms (create/update).
type GetAnnouncementFormDataResp struct {
	ID      uint   `json:"id"`      // ID
	Title   string `json:"title"`   // 标题名称
	Content string `json:"content"` // 内容
}

// UpdateAnnouncementReq represents announcement update data.
type UpdateAnnouncementReq struct {
	ID      uint   `uri:"id" binding:"required" swaggerignore:"true"` // ID
	Title   string `json:"title" binding:"required,min=1,max=100"`    // 名称
	Content string `json:"content" binding:"required"`                // 内容
	Status  string `json:"status" binding:"required"`                 // 状态
}

// UpdateAnnouncementResp represents announcement update response.
type UpdateAnnouncementResp struct {
}

// DeleteAnnouncementReq represents a request to delete a announcement.
type DeleteAnnouncementReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"` // ID
}

// DeleteAnnouncementResp represents announcement delete response.
type DeleteAnnouncementResp struct {
}

// BatchDeleteAnnouncementsReq represents a request to batch delete announcements.
type BatchDeleteAnnouncementsReq struct {
	IDs []uint `json:"ids" binding:"required,min=1,dive,required"` // ids
}

// BatchDeleteAnnouncementsResp represents batch delete announcements response.
type BatchDeleteAnnouncementsResp struct {
	SuccessCount int      `json:"success_count"` // 成功删除数量
	FailedCount  int      `json:"failed_count"`  // 失败数量
	FailedItems  []string `json:"failed_items"`  // 失败的信息（格式：名称: 错误原因）
}

// GetAnnouncementPublishersOptionsReq represents a request to get announcement publishers options.
type GetAnnouncementPublishersOptionsReq struct {
	SortReq
}

// GetAnnouncementPublishersOptionsResp represents announcement publishers options response.
type GetAnnouncementPublishersOptionsResp struct {
	Results []GetAnnouncementPublishersOptionsRespResult `json:"results"`
}

type GetAnnouncementPublishersOptionsRespResult struct {
	GetOptionsRespResult
}
