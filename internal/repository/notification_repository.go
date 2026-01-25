// Package repository provides data access implementations.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/internal/util"
	"github.com/xelarion/go-layout/pkg/errs"
)

var _ usecase.NotificationRepository = (*NotificationRepository)(nil)

// NotificationRepository is an implementation of the notification repository.
type NotificationRepository struct {
	data *Data
}

// NewNotificationRepository creates a new instance of notification repository.
func NewNotificationRepository(data *Data) usecase.NotificationRepository {
	return &NotificationRepository{
		data: data,
	}
}

// Create adds a new notification to the database.
func (r *NotificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	if err := r.data.DB(ctx).Create(notification).Error; err != nil {
		return errs.WrapInternal(err, "创建记录失败")
	}
	return nil
}

// List retrieves notifications with pagination and filtering.
func (r *NotificationRepository) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.Notification, int, error) {
	query := r.data.DB(ctx).Model(&model.Notification{}).
		Joins("INNER JOIN user_notifications ON user_notifications.notification_id = notifications.id")

	// 处理数据权限的 OR 条件（优先处理）
	if orConditions, ok := filters["_data_scope_or"].([]map[string]any); ok {
		query = ApplyORConditions(query, orConditions)
		delete(filters, "_data_scope_or") // 删除特殊标记，避免后续再次处理
	}

	for field, value := range filters {
		if value == nil {
			continue
		}

		switch field {
		case "ids":
			if ids, ok := value.([]uint); ok {
				query = query.Where("notifications.id IN (?)", ids)
			}
		case "key":
			if str, ok := value.(string); ok {
				query = query.Where("notifications.title LIKE @key", sql.Named("key", util.EscapeFullLike(str)))
			}
		case "type":
			if typ, ok := value.(enum.NotificationType); ok {
				query = query.Where("notifications.type = ?", typ)
			}
		case "user_id":
			if userId, ok := value.(uint); ok {
				query = query.Where("user_notifications.user_id = ?", userId)
			}
		case "is_read":
			if isRead, ok := value.(bool); ok {
				query = query.Where("user_notifications.is_read = ?", isRead)
			}
		default:
			query = query.Where("notifications."+field+" = ?", value)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "查询失败")
	}

	if sortClause != "" {
		query = query.Order(sortClause)
	} else {
		query = query.Order("notifications.id desc")
	}

	var notifications []*model.Notification
	if err := query.Limit(limit).Offset(offset).Find(&notifications).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "查询失败")
	}

	return notifications, int(total), nil
}

func (r *NotificationRepository) IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error) {
	return IsExists(ctx, r.data.DB(ctx), &model.Notification{}, filters, notFilters)
}

// Count counts the number of notifications.
func (r *NotificationRepository) Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error) {
	return Count(ctx, r.data.DB(ctx), &model.Notification{}, filters, notFilters)
}

// FindByID retrieves a notification by ID.
func (r *NotificationRepository) FindByID(ctx context.Context, id uint) (*model.Notification, error) {
	var notification model.Notification
	err := r.data.DB(ctx).First(&notification, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("记录不存在").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, errs.WrapInternal(err, "查询失败")
	}
	return &notification, nil
}

// FindByAnnouncementID 根据 announcement_id 查询通知
func (r *NotificationRepository) FindByAnnouncementID(ctx context.Context, announcementID uint) (*model.Notification, error) {
	var notification model.Notification
	err := r.data.DB(ctx).Where(fmt.Sprintf(`payload @> '{"announcement_id": %d}'`, announcementID)).First(&notification).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("记录不存在").
				WithReason(errs.ReasonNotFound).
				WithMeta("announcement_id", announcementID)
		}
		return nil, errs.WrapInternal(err, "查询失败")
	}
	return &notification, nil
}

// Update updates a notification.
func (r *NotificationRepository) Update(ctx context.Context, notification *model.Notification, params map[string]any) error {
	result := r.data.DB(ctx).Model(notification).Updates(params)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "更新失败")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("记录不存在").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", notification.ID)
	}

	return nil
}

// Delete removes notification by ID.
func (r *NotificationRepository) Delete(ctx context.Context, id uint) error {
	result := r.data.DB(ctx).Delete(&model.Notification{}, id)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "删除失败")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("记录不存在").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", id)
	}

	return nil
}
