package types

// Common notification-related types

// ListNotificationsReq represents notification list query parameters.
type ListNotificationsReq struct {
	PageReq
	SortReq
	Key      string `form:"key"`      // 标题
	Type     string `form:"type"`     // 类型
	Category string `form:"category"` // 类别(通知、消息)
	IsRead   *bool  `form:"is_read"`  // 是否已读
}

// ListNotificationsResp represents notification list response.
type ListNotificationsResp struct {
	Results    []*ListNotificationsRespResult `json:"results"`
	Pagination PageResp                       `json:"pagination"`
}

// ListNotificationsRespResult represents a single notification in the list.
type ListNotificationsRespResult struct {
	ID           uint   `json:"id"`            // ID
	Title        string `json:"title"`         // 标题
	Type         string `json:"type"`          // 类型
	TypeSlug     string `json:"type_slug"`     // 类型 slug
	Category     string `json:"category"`      // 类别(通知、消息)
	CategorySlug string `json:"category_slug"` // 类别 slug
	IsRead       bool   `json:"is_read"`       // 是否已读
	CreatedAt    string `json:"created_at"`    // 创建时间
}

type GetRecentNotificationsReq struct {
	IsRead *bool `form:"is_read"` // 是否已读
}

type GetRecentNotificationsResp struct {
	CategoryResults []*GetRecentNotificationsRespCategoryResult `json:"category_results"`
}

type GetRecentNotificationsRespCategoryResult struct {
	CategorySlug string                                            `json:"category_slug"` // 类别 slug
	Count        int                                               `json:"count"`         // 数量
	Results      []*GetRecentNotificationsRespCategoryResultResult `json:"results"`
}
type GetRecentNotificationsRespCategoryResultResult struct {
	ID           uint   `json:"id"`            // ID
	Title        string `json:"title"`         // 标题
	CategorySlug string `json:"category_slug"` // 类别 slug
	IsRead       bool   `json:"is_read"`       // 是否已读
	CreatedAt    string `json:"created_at"`    // 创建时间
}

// GetNotificationReq represents a request to get a specific notification.
type GetNotificationReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"` // ID
}

// GetNotificationResp represents a response for getting a notification.
type GetNotificationResp struct {
	ID        uint   `json:"id"`         // ID
	Title     string `json:"title"`      // 标题
	Content   string `json:"content"`    // 内容
	Type      string `json:"type"`       // 类型
	IsRead    bool   `json:"is_read"`    // 是否已读
	CreatedAt string `json:"created_at"` // 创建时间
}

// DeleteNotificationReq represents a request to delete a notification.
type DeleteNotificationReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"` // ID
}

// DeleteNotificationResp represents notification delete response.
type DeleteNotificationResp struct {
}

// BatchDeleteNotificationsReq represents a request to batch delete notifications.
type BatchDeleteNotificationsReq struct {
	IDs []uint `json:"ids" binding:"required,min=1,dive,required"` // ids
}

// BatchDeleteNotificationsResp represents batch delete notifications response.
type BatchDeleteNotificationsResp struct {
	SuccessCount int      `json:"success_count"` // 成功数量
	FailedCount  int      `json:"failed_count"`  // 失败数量
	FailedItems  []string `json:"failed_items"`  // 失败的信息（格式：名称: 错误原因）
}

// BatchReadNotificationsReq represents a request to batch read notifications.
type BatchReadNotificationsReq struct {
	IDs []uint `json:"ids" binding:"omitempty,min=1,dive,required"` // ids
}

// BatchReadNotificationsResp represents batch read notifications response.
type BatchReadNotificationsResp struct {
	SuccessCount int      `json:"success_count"` // 成功数量
	FailedCount  int      `json:"failed_count"`  // 失败数量
	FailedItems  []string `json:"failed_items"`  // 失败的信息（格式：名称: 错误原因）
}

// ReadAllNotificationsReq represents a request to read all notifications.
type ReadAllNotificationsReq struct {
	Category string `form:"category"` // 类别(为空时已读全部)
}

// ReadAllNotificationsResp represents read all notifications response.
type ReadAllNotificationsResp struct {
}

type GetTodoStatsReq struct {
}

type GetTodoStatsResp struct {
	Results []GetTodoStatsRespResult `json:"results"`
}

type GetTodoStatsRespResult struct {
	Type  string `json:"type"`  // 类型标识
	Name  string `json:"name"`  // 名称
	Count int    `json:"count"` // 数量
}
