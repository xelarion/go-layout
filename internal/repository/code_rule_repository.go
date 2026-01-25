// Package repository provides data access implementations.
package repository

import (
	"context"
	"database/sql"
	"errors"

	"gorm.io/gorm"

	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/internal/util"
	"github.com/xelarion/go-layout/pkg/errs"
)

var _ usecase.CodeRuleRepository = (*CodeRuleRepository)(nil)

// CodeRuleRepository is an implementation of the code_rule repository.
type CodeRuleRepository struct {
	data *Data
}

// NewCodeRuleRepository creates a new instance of code_rule repository.
func NewCodeRuleRepository(data *Data) usecase.CodeRuleRepository {
	return &CodeRuleRepository{
		data: data,
	}
}

// Create adds a new code_rule to the database.
func (r *CodeRuleRepository) Create(ctx context.Context, codeRule *model.CodeRule) error {
	if err := r.data.DB(ctx).Create(codeRule).Error; err != nil {
		return errs.WrapInternal(err, "failed to create code_rule")
	}
	return nil
}

// List retrieves code_rules with pagination and filtering.
func (r *CodeRuleRepository) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.CodeRule, int, error) {
	query := r.data.DB(ctx).Model(&model.CodeRule{})

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
		case "key":
			if str, ok := value.(string); ok {
				query = query.Where("code_rules.name LIKE @key OR prefix LIKE @key", sql.Named("key", util.EscapeFullLike(str)))
			}
		default:
			query = query.Where("code_rules."+field+" = ?", value)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to count code_rules")
	}

	if sortClause != "" {
		query = query.Order(sortClause)
	} else {
		query = query.Order("code_rules.id asc")
	}

	var codeRules []*model.CodeRule
	if err := query.Limit(limit).Offset(offset).Find(&codeRules).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to list code_rules")
	}

	return codeRules, int(total), nil
}

func (r *CodeRuleRepository) IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error) {
	return IsExists(ctx, r.data.DB(ctx), &model.CodeRule{}, filters, notFilters)
}

// Count counts the number of code_rules.
func (r *CodeRuleRepository) Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error) {
	return Count(ctx, r.data.DB(ctx), &model.CodeRule{}, filters, notFilters)
}

// FindByID retrieves a code_rule by ID.
func (r *CodeRuleRepository) FindByID(ctx context.Context, id uint) (*model.CodeRule, error) {
	var codeRule model.CodeRule
	err := r.data.DB(ctx).First(&codeRule, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("记录不存在").
				WithReason(errs.ReasonNotFound).
				WithMeta("id", id)
		}
		return nil, errs.WrapInternal(err, "查询失败")
	}
	return &codeRule, nil
}

// FindBySlug retrieves a code_rule by slug.
func (r *CodeRuleRepository) FindBySlug(ctx context.Context, slug string) (*model.CodeRule, error) {
	var codeRule model.CodeRule
	err := r.data.DB(ctx).First(&codeRule, "slug = ?", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("记录不存在").
				WithReason(errs.ReasonNotFound).
				WithMeta("slug", slug)
		}
		return nil, errs.WrapInternal(err, "查询失败")
	}
	return &codeRule, nil
}

// Update updates a code_rule.
func (r *CodeRuleRepository) Update(ctx context.Context, codeRule *model.CodeRule, params map[string]any) error {
	result := r.data.DB(ctx).Model(codeRule).Updates(params)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "更新失败")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("记录不存在").
			WithReason(errs.ReasonNotFound).
			WithMeta("id", codeRule.ID)
	}

	return nil
}

// Delete removes a code_rule by ID.
func (r *CodeRuleRepository) Delete(ctx context.Context, id uint) error {
	result := r.data.DB(ctx).Delete(&model.CodeRule{}, id)
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
