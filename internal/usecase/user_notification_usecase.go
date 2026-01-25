// Package usecase contains business logic.
package usecase

import (
	"context"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
)

// UserNotificationRepository defines methods for user_notification data access.
type UserNotificationRepository interface {
	Create(ctx context.Context, userNotification *model.UserNotification) error
	CreateInBatches(ctx context.Context, userNotifications []*model.UserNotification, batchSize int) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.UserNotification, int, error)
	IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error)
	Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error)
	FindByID(ctx context.Context, id uint) (*model.UserNotification, error)
	FindByNotificationIDAndUserID(ctx context.Context, notificationID uint, userID uint) (*model.UserNotification, error)
	Update(ctx context.Context, userNotification *model.UserNotification, params map[string]any) error
	Delete(ctx context.Context, id uint) error
	DeleteAllByNotificationID(ctx context.Context, notificationID uint) error
	ReadAll(ctx context.Context, userID uint, category *enum.NotificationCategory) error
}
