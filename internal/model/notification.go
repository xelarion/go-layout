package model

import (
	"gorm.io/datatypes"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model/gen"
)

type Notification struct {
	gen.Notification

	Type     enum.NotificationType                   `gorm:"column:type;type:smallint;not null;comment:类型" json:"type"`         // 类型
	Category enum.NotificationCategory               `gorm:"column:category;type:smallint;not null;comment:分类" json:"category"` // 分类
	Scope    enum.NotificationScope                  `gorm:"column:scope;type:smallint;not null;comment:范围" json:"scope"`       // 范围
	Payload  datatypes.JSONType[NotificationPayload] `gorm:"column:payload;type:jsonb;not null;comment:数据" json:"payload"`      // 数据
}

// NotificationPayload 通知附加数据
type NotificationPayload struct {
	UserIDs        []uint `json:"user_ids,omitempty"`        // Scope=enum.NotificationScopeUser 时的目标用户ID列表
	AnnouncementID uint   `json:"announcement_id,omitempty"` // Type=enum.NotificationTypeAnnouncement 时的目标公告ID
}

// GetResourceType 实现 Resource 接口
func (e *Notification) GetResourceType() string {
	return "notification"
}

// GetResourceID 实现 Resource 接口
func (e *Notification) GetResourceID() uint {
	return e.ID
}
