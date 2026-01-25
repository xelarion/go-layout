// Package repository provides data access implementations.
package repository

import (
	"context"
	"database/sql"
	"errors"

	"gorm.io/gorm"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/internal/util"
	"github.com/xelarion/go-layout/pkg/errs"
)

var _ usecase.UserNotificationRepository = (*UserNotificationRepository)(nil)

// UserNotificationRepository is an implementation of the user_notification repository.
type UserNotificationRepository struct {
	data *Data
}

// NewUserNotificationRepository creates a new instance of user_notification repository.
func NewUserNotificationRepository(data *Data) usecase.UserNotificationRepository {
	return &UserNotificationRepository{
		data: data,
	}
}

// Create adds a new user_notification to the database.
func (r *UserNotificationRepository) Create(ctx context.Context, userNotification *model.UserNotification) error {
	if err := r.data.DB(ctx).Create(userNotification).Error; err != nil {
		return errs.WrapInternal(err, "创建记录失败")
	}
	return nil
}

// List retrieves user_notifications with pagination and filtering.
func (r *UserNotificationRepository) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.UserNotification, int, error) {
	query := r.data.DB(ctx).Model(&model.UserNotification{})

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
				query = query.Where("user_notifications.id IN (?)", ids)
			}
		case "key":
			if str, ok := value.(string); ok {
				query = query.Where("user_notifications.name LIKE @key OR user_notifications.code LIKE @key", sql.Named("key", util.EscapeFullLike(str)))
			}
		default:
			query = query.Where("user_notifications."+field+" = ?", value)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "查询失败")
	}

	if sortClause != "" {
		query = query.Order(sortClause)
	} else {
		query = query.Order("user_notifications.id desc")
	}

	var userNotifications []*model.UserNotification
	if err := query.Limit(limit).Offset(offset).Find(&userNotifications).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "查询失败")
	}

	return userNotifications, int(total), nil
}

func (r *UserNotificationRepository) IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error) {
	return IsExists(ctx, r.data.DB(ctx), &model.UserNotification{}, filters, notFilters)
}

// Count counts the number of user_notifications.
func (r *UserNotificationRepository) Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error) {
	return Count(ctx, r.data.DB(ctx), &model.UserNotification{}, filters, notFilters)
}

// FindByID retrieves a user_notification by ID.
func (r *UserNotificationRepository) FindByID(ctx context.Context, id uint) (*model.UserNotification, error) {
	var userNotification model.UserNotification
	err := r.data.DB(ctx).First(&userNotification, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("记录不存在").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, errs.WrapInternal(err, "查询失败")
	}
	return &userNotification, nil
}

// FindByNotificationIDAndUserID find by notification_id and user_id
func (r *UserNotificationRepository) FindByNotificationIDAndUserID(ctx context.Context, notificationID uint, userID uint) (*model.UserNotification, error) {
	var userNotification model.UserNotification
	err := r.data.DB(ctx).Where("user_id = ? AND notification_id = ?", userID, notificationID).First(&userNotification).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("记录不存在").
				WithReason(errs.ReasonNotFound).
				WithMetaMap(map[string]any{
					"user_id":         userID,
					"notification_id": notificationID,
				})
		}
		return nil, errs.WrapInternal(err, "查询失败")
	}
	return &userNotification, nil
}

// Update updates a user_notification.
func (r *UserNotificationRepository) Update(ctx context.Context, userNotification *model.UserNotification, params map[string]any) error {
	result := r.data.DB(ctx).Model(userNotification).Updates(params)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "更新失败")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("记录不存在").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", userNotification.ID)
	}

	return nil
}

// Delete removes user_notification by ID.
func (r *UserNotificationRepository) Delete(ctx context.Context, id uint) error {
	result := r.data.DB(ctx).Delete(&model.UserNotification{}, id)
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

// DeleteAllByNotificationID 删除某个通知的所有用户通知
func (r *UserNotificationRepository) DeleteAllByNotificationID(ctx context.Context, notificationID uint) error {
	result := r.data.DB(ctx).Where("notification_id = ?", notificationID).Delete(&model.UserNotification{})
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "删除失败")
	}

	return nil
}

// ReadAll marks all user_notifications as read for a specific user.
// category 为 nil 时更新所有
func (r *UserNotificationRepository) ReadAll(ctx context.Context, userID uint, category *enum.NotificationCategory) error {
	tx := r.data.DB(ctx).Model(&model.UserNotification{}).
		Where("user_id = ?", userID).
		Where("is_read = false")

	if category != nil {
		tx = tx.Where("notification_id IN (?)",
			r.data.DB(ctx).Model(&model.Notification{}).
				Select("id").
				Where("category = ?", *category),
		)
	}

	result := tx.Updates(map[string]any{"is_read": true, "read_at": util.DBTimeNow()})
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "操作失败")
	}

	return nil
}

// CreateInBatches creates user_notifications in batches.
func (r *UserNotificationRepository) CreateInBatches(ctx context.Context, userNotifications []*model.UserNotification, batchSize int) error {
	if err := r.data.DB(ctx).CreateInBatches(userNotifications, batchSize).Error; err != nil {
		return errs.WrapInternal(err, "批量创建记录失败")
	}
	return nil
}
