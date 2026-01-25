package model

import (
	"github.com/xelarion/go-layout/internal/model/gen"
)

type UserNotification struct {
	gen.UserNotification
}

// GetResourceType 实现 Resource 接口
func (e *UserNotification) GetResourceType() string {
	return "user_notification"
}

// GetResourceID 实现 Resource 接口
func (e *UserNotification) GetResourceID() uint {
	return e.ID
}
