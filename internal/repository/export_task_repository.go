// Package repository provides data access implementations.
package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/errs"
)

var _ usecase.ExportTaskRepository = (*ExportTaskRepository)(nil)

// ExportTaskRepository is an implementation of the export_task repository.
type ExportTaskRepository struct {
	data *Data
}

// NewExportTaskRepository creates a new instance of export_task repository.
func NewExportTaskRepository(data *Data) usecase.ExportTaskRepository {
	return &ExportTaskRepository{
		data: data,
	}
}

// Create adds a new export_task to the database.
func (r *ExportTaskRepository) Create(ctx context.Context, exportTask *model.ExportTask) error {
	if err := r.data.DB(ctx).Create(exportTask).Error; err != nil {
		return errs.WrapInternal(err, "failed to create export_task")
	}
	return nil
}

// List retrieves export_tasks with pagination and filtering.
func (r *ExportTaskRepository) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.ExportTask, int, error) {
	query := r.data.DB(ctx).Model(&model.ExportTask{})

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
		case "completed_at_end":
			if val, ok := value.(time.Time); ok {
				query = query.Where("export_tasks.completed_at <= ?", val)
			}
		case "status_in":
			if val, ok := value.([]enum.ExportTaskStatus); ok {
				query = query.Where("export_tasks.status IN ?", val)
			}
		default:
			query = query.Where("export_tasks."+field+" = ?", value)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to count export_tasks")
	}

	if sortClause != "" {
		query = query.Order(sortClause)
	} else {
		query = query.Order("export_tasks.id DESC")
	}

	var exportTasks []*model.ExportTask
	if err := query.Limit(limit).Offset(offset).Find(&exportTasks).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to list export_tasks")
	}

	return exportTasks, int(total), nil
}

func (r *ExportTaskRepository) IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error) {
	return IsExists(ctx, r.data.DB(ctx), &model.ExportTask{}, filters, notFilters)
}

// Count counts the number of export_tasks.
func (r *ExportTaskRepository) Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error) {
	return Count(ctx, r.data.DB(ctx), &model.ExportTask{}, filters, notFilters)
}

// FindByID retrieves a export_task by ID.
func (r *ExportTaskRepository) FindByID(ctx context.Context, id uint) (*model.ExportTask, error) {
	var exportTask model.ExportTask
	err := r.data.DB(ctx).First(&exportTask, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("记录不存在").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, errs.WrapInternal(err, "查询失败")
	}
	return &exportTask, nil
}

// FindByIDForUpdate retrieves a export_task by ID with FOR UPDATE lock.
// 必须在事务中使用,用于防止并发修改
func (r *ExportTaskRepository) FindByIDForUpdate(ctx context.Context, id uint) (*model.ExportTask, error) {
	var exportTask model.ExportTask
	err := r.data.DB(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&exportTask, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("记录不存在").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, errs.WrapInternal(err, "查询失败")
	}
	return &exportTask, nil
}

// Update updates a export_task.
func (r *ExportTaskRepository) Update(ctx context.Context, exportTask *model.ExportTask, params map[string]any) error {
	result := r.data.DB(ctx).Model(exportTask).Updates(params)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "更新失败")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("记录不存在").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", exportTask.ID)
	}

	return nil
}

// Delete removes a export_task by ID.
func (r *ExportTaskRepository) Delete(ctx context.Context, id uint) error {
	result := r.data.DB(ctx).Delete(&model.ExportTask{}, id)
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
