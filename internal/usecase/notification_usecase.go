// Package usecase contains business logic.
package usecase

import (
	"context"
	"fmt"

	"gorm.io/datatypes"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/model/gen"
	"github.com/xelarion/go-layout/internal/util"
	"github.com/xelarion/go-layout/pkg/errs"
)

// Notification contains notification data
type Notification struct {
	*model.Notification
	IsRead bool
}

// CreateNotificationParams contains all parameters needed to create a notification
type CreateNotificationParams struct {
	CreatedBy      uint // Creator user ID
	Title          string
	Content        string
	Type           enum.NotificationType
	Category       enum.NotificationCategory
	Scope          enum.NotificationScope
	UserIDs        []uint // 仅当 Scope 为 enum.NotificationScopeUser 时有用
	AnnouncementID uint   // 仅当 Type 为 enum.NotificationTypeAnnouncement 时有用
}

// BatchDeleteNotificationsResult represents the result of a single notification deletion.
type BatchDeleteNotificationsResult = BatchItemCommonResult

// BatchReadNotificationsResult represents the result of a single notification deletion.
type BatchReadNotificationsResult = BatchItemCommonResult

// TodoStat represents a todo statistic entry
type TodoStat struct {
	Type  enum.TodoStatType
	Name  string
	Count int
}

// NotificationRepository defines methods for notification data access.
type NotificationRepository interface {
	Create(ctx context.Context, notification *model.Notification) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.Notification, int, error)
	IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error)
	Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error)
	FindByID(ctx context.Context, id uint) (*model.Notification, error)
	FindByAnnouncementID(ctx context.Context, announcementID uint) (*model.Notification, error)
	Update(ctx context.Context, notification *model.Notification, params map[string]any) error
	Delete(ctx context.Context, id uint) error
}

// NotificationUseCase implements business logic for notification operations.
type NotificationUseCase struct {
	tx                   Transaction
	amqpMessageUseCase   *AmqpMessageUseCase
	notificationRepo     NotificationRepository
	userNotificationRepo UserNotificationRepository
}

// NewNotificationUseCase creates a new instance of NotificationUseCase.
func NewNotificationUseCase(
	tx Transaction,
	amqpMessageUseCase *AmqpMessageUseCase,
	repo NotificationRepository,
	userNotificationRepo UserNotificationRepository,
) *NotificationUseCase {
	return &NotificationUseCase{
		tx:                   tx,
		amqpMessageUseCase:   amqpMessageUseCase,
		notificationRepo:     repo,
		userNotificationRepo: userNotificationRepo,
	}
}

// Create creates a new notification.
func (uc *NotificationUseCase) Create(ctx context.Context, params CreateNotificationParams) (uint, error) {
	// 构建 payload
	payload := model.NotificationPayload{}
	if params.Scope == enum.NotificationScopeUser {
		if len(params.UserIDs) == 0 {
			return 0, errs.NewBusiness("用户通知必须指定用户ID")
		}
		payload.UserIDs = params.UserIDs
	}

	if params.Type == enum.NotificationTypeAnnouncement {
		if params.AnnouncementID == 0 {
			return 0, errs.NewBusiness("公告通知必须指定公告ID")
		}
		payload.AnnouncementID = params.AnnouncementID
	}

	notification := &model.Notification{
		Notification: gen.Notification{
			CreatedBy: params.CreatedBy,
			Title:     params.Title,
			Content:   params.Content,
		},
		Type:     params.Type,
		Category: params.Category,
		Scope:    params.Scope,
		Payload:  datatypes.NewJSONType(payload),
	}

	err := uc.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := uc.notificationRepo.Create(ctx, notification); err != nil {
			return err
		}

		// 统一通过队列异步创建 user_notifications
		if err := uc.amqpMessageUseCase.PublishCreateUserNotification(ctx, notification.ID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return notification.ID, nil
}

// List returns a list of notifications with pagination and filtering.
func (uc *NotificationUseCase) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string, userID uint) ([]*Notification, int, error) {
	notifications, count, err := uc.notificationRepo.List(ctx, filters, limit, offset, sortClause)
	if err != nil {
		return nil, 0, err
	}

	records := make([]*Notification, 0, len(notifications))
	for _, notification := range notifications {
		userNotification, err := uc.userNotificationRepo.FindByNotificationIDAndUserID(ctx, notification.ID, userID)
		if err != nil {
			return nil, 0, err
		}

		record := &Notification{
			Notification: notification,
			IsRead:       userNotification.IsRead,
		}
		records = append(records, record)
	}

	return records, count, nil
}

// GetByID retrieves notification by ID.
func (uc *NotificationUseCase) GetByID(ctx context.Context, id uint, userID uint) (*Notification, error) {
	notification, err := uc.notificationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	userNotification, err := uc.userNotificationRepo.FindByNotificationIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	return &Notification{
		Notification: notification,
		IsRead:       userNotification.IsRead,
	}, nil
}

// GetAndRead retrieves notification by ID and marks it as read.
func (uc *NotificationUseCase) GetAndRead(ctx context.Context, id uint, userID uint) (*Notification, error) {
	notification, err := uc.notificationRepo.FindByID(ctx, id)
	if err != nil {
		if errs.IsReason(err, errs.ReasonNotFound) {
			return nil, errs.NewBusiness("该内容不存在/已删除").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, err
	}

	userNotification, err := uc.userNotificationRepo.FindByNotificationIDAndUserID(ctx, id, userID)
	if err != nil {
		if errs.IsReason(err, errs.ReasonNotFound) {
			return nil, errs.NewBusiness("该内容不存在/已删除").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, err
	}

	err = uc.read(ctx, userNotification)
	if err != nil {
		return nil, err
	}

	return &Notification{
		Notification: notification,
		IsRead:       true,
	}, nil
}

// Delete removes a notification.
func (uc *NotificationUseCase) Delete(ctx context.Context, id uint, userID uint) error {
	notification, err := uc.notificationRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	userNotification, err := uc.userNotificationRepo.FindByNotificationIDAndUserID(ctx, id, userID)
	if err != nil {
		return err
	}

	return uc.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := uc.userNotificationRepo.Delete(ctx, userNotification.ID); err != nil {
			return err
		}

		hasUserNotification, err := uc.userNotificationRepo.IsExists(ctx, map[string]any{"notification_id": notification.ID}, nil)
		if err != nil {
			return err
		}

		if !hasUserNotification {
			if err := uc.notificationRepo.Delete(ctx, id); err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchDelete removes multiple notifications and returns detailed results.
func (uc *NotificationUseCase) BatchDelete(ctx context.Context, ids []uint, userID uint) []BatchDeleteNotificationsResult {
	results := make([]BatchDeleteNotificationsResult, 0, len(ids))

	for _, id := range ids {
		result := BatchDeleteNotificationsResult{
			ID:      id,
			Success: false,
		}

		notification, err := uc.notificationRepo.FindByID(ctx, id)
		if err != nil {
			result.Identifier = fmt.Sprintf("ID %d 未知记录", id)
			result.Error = err
			results = append(results, result)
			continue
		}

		result.Identifier = notification.Title

		err = uc.Delete(ctx, id, userID)
		if err != nil {
			result.Error = err
		} else {
			result.Success = true
		}

		results = append(results, result)
	}

	return results
}

// DeleteByAnnouncementID 删除公告的所有通知
func (uc *NotificationUseCase) DeleteByAnnouncementID(ctx context.Context, announcementID uint) error {
	return uc.tx.Transaction(ctx, func(ctx context.Context) error {
		notification, err := uc.notificationRepo.FindByAnnouncementID(ctx, announcementID)
		if err != nil {
			if errs.IsReason(err, errs.ReasonNotFound) {
				return nil
			}
			return err
		}

		if err := uc.notificationRepo.Delete(ctx, notification.ID); err != nil {
			if !errs.IsReason(err, errs.ReasonNotFound) {
				return err
			}
		}

		return uc.userNotificationRepo.DeleteAllByNotificationID(ctx, notification.ID)
	})
}

// Read marks a notification as read.
func (uc *NotificationUseCase) Read(ctx context.Context, id uint, userID uint) error {
	userNotification, err := uc.userNotificationRepo.FindByNotificationIDAndUserID(ctx, id, userID)
	if err != nil {
		return err
	}

	return uc.read(ctx, userNotification)
}

// BatchRead marks multiple notifications as read and returns detailed results.
func (uc *NotificationUseCase) BatchRead(ctx context.Context, ids []uint, userID uint) []BatchReadNotificationsResult {
	results := make([]BatchReadNotificationsResult, 0, len(ids))

	for _, id := range ids {
		result := BatchReadNotificationsResult{
			ID:      id,
			Success: false,
		}

		notification, err := uc.notificationRepo.FindByID(ctx, id)
		if err != nil {
			result.Identifier = fmt.Sprintf("ID %d 未知记录", id)
			result.Error = err
			results = append(results, result)
			continue
		}

		result.Identifier = notification.Title

		err = uc.Read(ctx, id, userID)
		if err != nil {
			result.Error = err
		} else {
			result.Success = true
		}

		results = append(results, result)
	}

	return results
}

// ReadAll marks all user_notifications as read for a specific user.
func (uc *NotificationUseCase) ReadAll(ctx context.Context, userID uint, category *enum.NotificationCategory) error {
	return uc.userNotificationRepo.ReadAll(ctx, userID, category)
}

// read marks a notification as read.
func (uc *NotificationUseCase) read(ctx context.Context, userNotification *model.UserNotification) error {
	if userNotification.IsRead {
		return nil
	}

	return uc.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := uc.userNotificationRepo.Update(ctx, userNotification, map[string]any{
			"is_read": true,
			"read_at": util.DBTimeNow(),
		}); err != nil {
			return err
		}
		return nil
	})
}

// GetTodoStats gets todo statistics for a user
func (uc *NotificationUseCase) GetTodoStats(ctx context.Context, userID uint) ([]*TodoStat, error) {
	stats := make([]*TodoStat, 0)

	// TODO: Implement actual statistics logic
	// For now, return example slice
	stats = append(stats, &TodoStat{
		Type:  enum.TodoStatTypeWorkOrder,
		Name:  enum.TodoStatTypeText[enum.TodoStatTypeWorkOrder],
		Count: 5,
	})
	stats = append(stats, &TodoStat{
		Type:  enum.TodoStatTypeProductionReport,
		Name:  enum.TodoStatTypeText[enum.TodoStatTypeProductionReport],
		Count: 3,
	})

	return stats, nil
}
