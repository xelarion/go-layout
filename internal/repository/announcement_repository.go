// Package repository provides data access implementations.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/internal/util"
	"github.com/xelarion/go-layout/pkg/errs"
)

var _ usecase.AnnouncementRepository = (*AnnouncementRepository)(nil)

// AnnouncementRepository is an implementation of the announcement repository.
type AnnouncementRepository struct {
	data *Data
}

// NewAnnouncementRepository creates a new instance of announcement repository.
func NewAnnouncementRepository(data *Data) usecase.AnnouncementRepository {
	return &AnnouncementRepository{
		data: data,
	}
}

// Create adds a new announcement to the database.
func (r *AnnouncementRepository) Create(ctx context.Context, announcement *model.Announcement) error {
	if err := r.data.DB(ctx).Create(announcement).Error; err != nil {
		return errs.WrapInternal(err, "创建记录失败")
	}
	return nil
}

// List retrieves announcements with pagination and filtering.
func (r *AnnouncementRepository) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.Announcement, int, error) {
	query := r.data.DB(ctx).Model(&model.Announcement{})

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
				query = query.Where("announcements.id IN (?)", ids)
			}
		case "key":
			if str, ok := value.(string); ok {
				query = query.Where("announcements.title LIKE @key", sql.Named("key", util.EscapeFullLike(str)))
			}
		case "published_at_start":
			if val, ok := value.(time.Time); ok {
				query = query.Where("announcements.published_at >= ?", val)
			}
		case "published_at_end":
			if val, ok := value.(time.Time); ok {
				query = query.Where("announcements.published_at <= ?", val)
			}
		default:
			query = query.Where("announcements."+field+" = ?", value)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "查询失败")
	}

	if sortClause != "" {
		query = query.Order(sortClause)
	} else {
		query = query.Order("announcements.id desc")
	}

	var announcements []*model.Announcement
	if err := query.Limit(limit).Offset(offset).Find(&announcements).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "查询失败")
	}

	return announcements, int(total), nil
}

func (r *AnnouncementRepository) IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error) {
	return IsExists(ctx, r.data.DB(ctx), &model.Announcement{}, filters, notFilters)
}

// Count counts the number of announcements.
func (r *AnnouncementRepository) Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error) {
	return Count(ctx, r.data.DB(ctx), &model.Announcement{}, filters, notFilters)
}

// FindByID retrieves a announcement by ID.
func (r *AnnouncementRepository) FindByID(ctx context.Context, id uint) (*model.Announcement, error) {
	var announcement model.Announcement
	err := r.data.DB(ctx).First(&announcement, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("记录不存在").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, errs.WrapInternal(err, "查询失败")
	}
	return &announcement, nil
}

// FindByIDForUpdate retrieves a batch_record by ID with FOR UPDATE lock.
// 必须在事务中使用,用于防止并发修改
func (r *AnnouncementRepository) FindByIDForUpdate(ctx context.Context, id uint) (*model.Announcement, error) {
	var announcement model.Announcement
	err := r.data.DB(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&announcement, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("记录不存在").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, errs.WrapInternal(err, "查询失败")
	}
	return &announcement, nil
}

// Update updates a announcement.
func (r *AnnouncementRepository) Update(ctx context.Context, announcement *model.Announcement, params map[string]any) error {
	result := r.data.DB(ctx).Model(announcement).Updates(params)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "更新失败")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("记录不存在").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", announcement.ID)
	}

	return nil
}

// Delete removes announcement by ID.
func (r *AnnouncementRepository) Delete(ctx context.Context, id uint) error {
	result := r.data.DB(ctx).Delete(&model.Announcement{}, id)
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

func (r *AnnouncementRepository) ListPublishers(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*usecase.AnnouncementPublisher, int, error) {
	query := r.data.DB(ctx).Model(&model.Announcement{}).
		Where("announcements.publisher_id IS NOT NULL").
		Select("DISTINCT announcements.publisher_id as id, users.full_name AS full_name").
		Joins("LEFT JOIN users on users.id = announcements.publisher_id AND users.deleted_at = 0")

		// 处理数据权限的 OR 条件（优先处理）
	if orConditions, ok := filters["_data_scope_or"].([]map[string]any); ok {
		query = ApplyORConditions(query, orConditions)
		delete(filters, "_data_scope_or") // 删除特殊标记，避免后续再次处理
	}

	// 处理其他过滤条件
	for field, value := range filters {
		if value == nil {
			continue
		}

		switch field {
		default:
			query = query.Where("announcements."+field+" = ?", value)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "查询发布者数量失败")
	}

	if sortClause != "" {
		query = query.Order(sortClause)
	} else {
		query = query.Order("id desc")
	}

	var publishers []*usecase.AnnouncementPublisher
	if err := query.Limit(limit).Offset(offset).Find(&publishers).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "查询发布者失败")
	}

	return publishers, int(total), nil
}
