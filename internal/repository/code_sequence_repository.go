// Package repository provides data access implementations.
package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/errs"
)

var _ usecase.CodeSequenceRepository = (*CodeSequenceRepository)(nil)

// CodeSequenceRepository is an implementation of the code_sequence repository.
type CodeSequenceRepository struct {
	data *Data
}

// NewCodeSequenceRepository creates a new instance of code_sequence repository.
func NewCodeSequenceRepository(data *Data) usecase.CodeSequenceRepository {
	return &CodeSequenceRepository{
		data: data,
	}
}

// Create adds a new code_sequence to the database.
func (r *CodeSequenceRepository) Create(ctx context.Context, codeSequence *model.CodeSequence) error {
	if err := r.data.DB(ctx).Create(codeSequence).Error; err != nil {
		return errs.WrapInternal(err, "failed to create code_sequence")
	}
	return nil
}

// List retrieves code_sequences with pagination and filtering.
func (r *CodeSequenceRepository) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.CodeSequence, int, error) {
	query := r.data.DB(ctx).Model(&model.CodeSequence{})

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
				query = query.Where("code_sequences.created_at >= ?", val)
			}
		case "created_at_end":
			if val, ok := value.(time.Time); ok {
				query = query.Where("code_sequences.created_at <= ?", val)
			}
		default:
			query = query.Where("code_sequences."+field+" = ?", value)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to count code_sequences")
	}

	if sortClause != "" {
		query = query.Order(sortClause)
	} else {
		query = query.Order("code_sequences.id DESC")
	}

	var codeSequences []*model.CodeSequence
	if err := query.Limit(limit).Offset(offset).Find(&codeSequences).Error; err != nil {
		return nil, 0, errs.WrapInternal(err, "failed to list code_sequences")
	}

	return codeSequences, int(total), nil
}

func (r *CodeSequenceRepository) IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error) {
	return IsExists(ctx, r.data.DB(ctx), &model.CodeSequence{}, filters, notFilters)
}

// Count counts the number of code_sequences.
func (r *CodeSequenceRepository) Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error) {
	return Count(ctx, r.data.DB(ctx), &model.CodeSequence{}, filters, notFilters)
}

// FindByRuleSlugAndTimeKey 通过复合主键查询序列记录
func (r *CodeSequenceRepository) FindByRuleSlugAndTimeKey(ctx context.Context, ruleSlug, timeKey string) (*model.CodeSequence, error) {
	var sequence model.CodeSequence
	err := r.data.DB(ctx).Where("rule_slug = ? AND time_key = ?", ruleSlug, timeKey).First(&sequence).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewBusiness("序列记录不存在").
				WithReason(errs.ReasonNotFound).
				WithMeta("rule_slug", ruleSlug).
				WithMeta("time_key", timeKey)
		}
		return nil, errs.WrapInternal(err, "查询序列记录失败")
	}
	return &sequence, nil
}

// UpdateByKey 通过复合主键更新序列记录
func (r *CodeSequenceRepository) UpdateByKey(ctx context.Context, ruleSlug, timeKey string, params map[string]any) error {
	result := r.data.DB(ctx).Model(&model.CodeSequence{}).
		Where("rule_slug = ? AND time_key = ?", ruleSlug, timeKey).
		Updates(params)
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "更新失败")
	}

	if result.RowsAffected == 0 {
		return errs.NewBusiness("记录不存在").
			WithReason(errs.ReasonNotFound).
			WithMeta("rule_slug", ruleSlug).
			WithMeta("time_key", timeKey)
	}

	return nil
}

// DeleteByKey 通过复合主键删除序列记录
func (r *CodeSequenceRepository) DeleteByKey(ctx context.Context, ruleSlug, timeKey string) error {
	result := r.data.DB(ctx).Where("rule_slug = ? AND time_key = ?", ruleSlug, timeKey).
		Delete(&model.CodeSequence{})
	if result.Error != nil {
		return errs.WrapInternal(result.Error, "删除失败")
	}
	if result.RowsAffected == 0 {
		return errs.NewBusiness("记录不存在").
			WithReason(errs.ReasonNotFound).
			WithMeta("rule_slug", ruleSlug).
			WithMeta("time_key", timeKey)
	}
	return nil
}

// GetNextSerial 原子性地获取下一个序列号（使用PostgreSQL的INSERT ON CONFLICT）
// 使用 UPSERT 模式保证并发安全，并在数据库层面检查最大值限制：
//   - 如果记录不存在：插入初始值
//   - 如果记录存在：递增 current_serial，并兼容编码规则调整
//   - 如果递增后超过最大值：返回业务错误
//
// 参数:
//   - ruleSlug: 编码规则标识
//   - timeKey: 时间分组键
//   - initialSerial: 初始序列号（首次插入时使用，规则调整时也会应用）
//   - step: 递增步长
//   - maxSerial: 最大序列号限制
//
// 返回:
//   - 递增后的新序列号
//   - 如果超过最大值限制，返回业务错误
//
// 兼容场景:
//  1. 正常递增：current=10, initial=1, step=1, max=999 → next=11 ✓
//  2. 规则调整（初始值提高）：current=10, initial=20, step=1, max=999 → next=20 ✓
//  3. 达到上限：current=998, initial=1, step=1, max=999 → next=999 ✓
//  4. 超过上限：current=999, initial=1, step=1, max=999 → 业务错误 ✗
//  5. 首次插入超限：initial=1000, max=999 → 业务错误 ✗
//  6. 规则调整后超限：current=990, initial=995, step=1, max=999 → next=995...999，最后报错 ✗
//
// 必须在事务中调用
func (r *CodeSequenceRepository) GetNextSerial(
	ctx context.Context,
	ruleSlug, timeKey string,
	initialSerial, step, maxSerial int64,
) (int64, error) {
	db := r.data.DB(ctx)

	// 使用原生SQL实现 INSERT ... ON CONFLICT ... RETURNING
	// 这是PostgreSQL的原子操作，完全并发安全
	//
	// SQL逻辑（简化版）：
	//   1. 首次插入：仅在 initialSerial <= maxSerial 时插入
	//   2. 后续更新：
	//      - 计算 next_serial = GREATEST(current_serial + step, initialSerial)
	//      - 使用 GREATEST 兼容编码规则调整（initialSerial提高的情况）
	//      - 仅在 next_serial <= maxSerial 时成功更新
	//   3. 如果超过最大值：INSERT WHERE 或 DO UPDATE WHERE 条件不满足，返回空结果
	//
	// 性能考虑：
	//   - 唯一索引 (rule_slug, time_key) 确保高效查询和冲突检测
	//   - RETURNING 子句一次性返回结果，无需额外查询
	//   - GREATEST 函数在数据库层执行，高效且原子
	//
	// 并发安全性：
	//   - INSERT ON CONFLICT 是PostgreSQL的行级锁，自动处理并发
	//   - 多个并发请求会排队执行，每次递增互不影响
	//   - 事务隔离级别保证序列号不重复、不跳号（除非回滚）
	var newSerial int64

	// 注意：
	//   1. 使用 INSERT ... SELECT ... WHERE 在首次插入时也检查最大值
	//   2. 使用 CAST 显式指定类型，避免 PostgreSQL "operator is not unique" 错误
	//   3. 如果返回空结果（RowsAffected=0），说明超过最大值限制
	result := db.Raw(`
		INSERT INTO code_sequences (rule_slug, time_key, current_serial, created_at)
		SELECT 
			@rule_slug,
			@time_key,
			CAST(@initial_serial AS BIGINT),
			NOW()
		WHERE CAST(@initial_serial AS BIGINT) <= CAST(@max_serial AS BIGINT)
		ON CONFLICT (rule_slug, time_key)
		DO UPDATE SET 
			current_serial = GREATEST(
				code_sequences.current_serial + CAST(@step AS BIGINT),
				CAST(@initial_serial AS BIGINT)
			)
		WHERE GREATEST(
			code_sequences.current_serial + CAST(@step AS BIGINT),
			CAST(@initial_serial AS BIGINT)
		) <= CAST(@max_serial AS BIGINT)
		RETURNING current_serial
	`, map[string]interface{}{
		"rule_slug":      ruleSlug,
		"time_key":       timeKey,
		"initial_serial": initialSerial,
		"step":           step,
		"max_serial":     maxSerial,
	}).Scan(&newSerial)

	if result.Error != nil {
		return 0, errs.WrapInternal(result.Error, "获取序列号失败")
	}

	// 如果没有返回行（RowsAffected=0），说明超过了最大值限制
	if result.RowsAffected == 0 {
		return 0, errs.NewBusiness("流水号已超出最大长度限制，请调整编码规则以支持更大的流水号范围").
			WithReason(errs.ReasonInvalidState).
			WithMeta("max_serial", maxSerial).
			WithMeta("initial_serial", initialSerial)
	}

	return newSerial, nil
}
