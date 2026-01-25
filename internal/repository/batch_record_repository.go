// Package repository provides data access implementations.
package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/errs"
)

var _ usecase.BatchRecordRepository = (*BatchRecordRepository)(nil)

// BatchRecordRepository is an implementation of the batch_record repository.
type BatchRecordRepository struct {
	data *Data
}

// NewBatchRecordRepository creates a new instance of batch_record repository.
func NewBatchRecordRepository(data *Data) usecase.BatchRecordRepository {
	return &BatchRecordRepository{
		data: data,
	}
}

// Create adds a new batch_record to the database.
func (r *BatchRecordRepository) Create(ctx context.Context, batchRecord *model.BatchRecord) error {
	if err := r.data.DB(ctx).Create(batchRecord).Error; err != nil {
		return errs.WrapInternal(err, "failed to create batch_record")
	}
	return nil
}

// List retrieves batch_records with pagination and filtering.
func (r *BatchRecordRepository) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.BatchRecord, int, error) {
	query := r.data.DB(ctx).Model(&model.BatchRecord{})

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
		case "created_at_start":
			if val, ok := value.(time.Time); ok {
				query = query.Where("batch_records.created_at >= ?", val)
			}
		case "created_at_end":
			if val, ok := value.(time.Time); ok {
				query = query.Where("batch_records.created_at <= ?", val)
			}
		default:
			query = query.Where("batch_records."+field+" = ?", value)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to count batch_records")
	}

	if sortClause != "" {
		query = query.Order(sortClause)
	} else {
		query = query.Order("batch_records.id DESC")
	}

	var batchRecords []*model.BatchRecord
	if err := query.Limit(limit).Offset(offset).Find(&batchRecords).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to list batch_records")
	}

	return batchRecords, int(total), nil
}

func (r *BatchRecordRepository) IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error) {
	return IsExists(ctx, r.data.DB(ctx), &model.BatchRecord{}, filters, notFilters)
}

// Count counts the number of batch_records.
func (r *BatchRecordRepository) Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error) {
	return Count(ctx, r.data.DB(ctx), &model.BatchRecord{}, filters, notFilters)
}

// FindByID retrieves a batch_record by ID.
func (r *BatchRecordRepository) FindByID(ctx context.Context, id uint) (*model.BatchRecord, error) {
	var batchRecord model.BatchRecord
	err := r.data.DB(ctx).First(&batchRecord, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("记录不存在").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, errs.WrapInternal(err, "查询失败")
	}
	return &batchRecord, nil
}

// FindByIDForUpdate retrieves a batch_record by ID with FOR UPDATE lock.
// 必须在事务中使用,用于防止并发修改
func (r *BatchRecordRepository) FindByIDForUpdate(ctx context.Context, id uint) (*model.BatchRecord, error) {
	var batchRecord model.BatchRecord
	err := r.data.DB(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&batchRecord, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("记录不存在").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, errs.WrapInternal(err, "查询失败")
	}
	return &batchRecord, nil
}

// Update updates a batch_record.
func (r *BatchRecordRepository) Update(ctx context.Context, batchRecord *model.BatchRecord, params map[string]any) error {
	result := r.data.DB(ctx).Model(batchRecord).Updates(params)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "更新失败")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("记录不存在").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", batchRecord.ID)
	}

	return nil
}

// Delete removes a batch_record by ID.
func (r *BatchRecordRepository) Delete(ctx context.Context, id uint) error {
	result := r.data.DB(ctx).Delete(&model.BatchRecord{}, id)
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
