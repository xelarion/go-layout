package usecase

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/util"
	"github.com/xelarion/go-layout/pkg/errs"
)

// CodeSequenceRepository 序列号仓储接口
type CodeSequenceRepository interface {
	Create(ctx context.Context, codeSequence *model.CodeSequence) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.CodeSequence, int, error)
	IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error)
	Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error)
	FindByRuleSlugAndTimeKey(ctx context.Context, ruleSlug, timeKey string) (*model.CodeSequence, error)
	UpdateByKey(ctx context.Context, ruleSlug, timeKey string, params map[string]any) error
	DeleteByKey(ctx context.Context, ruleSlug, timeKey string) error
	// GetNextSerial 原子性地获取下一个序列号
	// 在数据库层面检查最大值限制，确保并发安全。必须在事务中调用。
	GetNextSerial(ctx context.Context, ruleSlug, timeKey string, initialSerial, step, maxSerial int64) (int64, error)
}

// CodeGeneratorUseCase 编码生成器
// 负责根据编码规则生成唯一编码，保证并发安全
type CodeGeneratorUseCase struct {
	tx               Transaction
	codeRuleRepo     CodeRuleRepository
	codeSequenceRepo CodeSequenceRepository
}

// NewCodeGeneratorUseCase 创建编码生成器实例
func NewCodeGeneratorUseCase(
	tx Transaction,
	codeRuleRepo CodeRuleRepository,
	codeSequenceRepo CodeSequenceRepository,
) *CodeGeneratorUseCase {
	return &CodeGeneratorUseCase{
		tx:               tx,
		codeRuleRepo:     codeRuleRepo,
		codeSequenceRepo: codeSequenceRepo,
	}
}

// GenerateCode 生成编码
// 使用数据库事务和行级锁保证并发安全，多个并发请求会串行执行序列号递增
// 返回格式：prefix + timePart + serialStr，例如：PO2024010100001
func (uc *CodeGeneratorUseCase) GenerateCode(ctx context.Context, ruleSlug string) (string, error) {
	rule, err := uc.codeRuleRepo.FindBySlug(ctx, ruleSlug)
	if err != nil {
		return "", err
	}

	// 构建时间部分和时间键
	timePart := uc.buildTimePart(rule.TimeRule)
	timeKey := uc.buildTimeKey(rule.GenerateType)

	// 计算最大序列号并验证初始值
	maxSerial := uc.calculateMaxSerial(rule.SerialLength)
	initialSerial := int64(rule.InitialSerial)
	if initialSerial > maxSerial {
		return "", errs.NewBusiness("初始流水号超过最大值限制，请调整编码规则").
			WithReason(errs.ReasonBadRequest).
			WithMeta("initial_serial", initialSerial).
			WithMeta("max_serial", maxSerial).
			WithMeta("serial_length", rule.SerialLength)
	}

	// 在事务中原子递增流水号
	var newSerial int64
	err = uc.tx.Transaction(ctx, func(txCtx context.Context) error {
		// Repository层：原子性地获取下一个序列号，并在数据库层面检查最大值
		var innerErr error
		newSerial, innerErr = uc.codeSequenceRepo.GetNextSerial(
			txCtx,
			rule.Slug,
			timeKey,
			initialSerial,
			int64(rule.Step),
			maxSerial,
		)
		return innerErr
	})
	if err != nil {
		return "", err
	}

	// 格式化流水号
	serialStr := uc.formatSerial(newSerial, int(rule.SerialLength))
	// 组装最终编码
	code := rule.Prefix + timePart + serialStr
	return code, nil
}

// PreviewCode 预览编码（不消耗序列号）
func (uc *CodeGeneratorUseCase) PreviewCode(
	ctx context.Context,
	prefix string,
	timeRule enum.CodeRuleTimeRule,
	generateType enum.CodeRuleGenerateType,
	initialSerial int32,
	serialLength int32,
	step int32,
) (firstCode string, endCode string, err error) {
	// 构建时间部分
	timePart := uc.buildTimePart(timeRule)

	// 计算最大流水号并验证初始值
	maxSerial := uc.calculateMaxSerial(serialLength)
	if int64(initialSerial) > maxSerial {
		return "", "", errs.NewBusiness("初始流水号超过最大值限制").
			WithReason(errs.ReasonBadRequest).
			WithMeta("initial_serial", initialSerial).
			WithMeta("max_serial", maxSerial).
			WithMeta("serial_length", serialLength)
	}

	// 第一个编码：使用初始流水号
	firstCode = prefix + timePart + uc.formatSerial(int64(initialSerial), int(serialLength))

	// 使用数学公式计算最后一个流水号
	// lastSerial = initialSerial + floor((maxSerial - initialSerial) / step) * step
	lastSerial := int64(initialSerial) + ((maxSerial-int64(initialSerial))/int64(step))*int64(step)

	endCode = prefix + timePart + uc.formatSerial(lastSerial, int(serialLength))

	return firstCode, endCode, nil
}

// buildTimePart 构建时间部分
func (uc *CodeGeneratorUseCase) buildTimePart(timeRule enum.CodeRuleTimeRule) string {
	if timeRule == enum.CodeRuleTimeRuleNone {
		return ""
	}

	now := time.Now().In(util.TimeLocal)

	switch timeRule {
	case enum.CodeRuleTimeRuleYear:
		return now.Format("2006")
	case enum.CodeRuleTimeRuleYearMonth:
		return now.Format("200601")
	case enum.CodeRuleTimeRuleYearMonthDay:
		return now.Format("20060102")
	case enum.CodeRuleTimeRuleYearMonthDayHour:
		return now.Format("2006010215")
	case enum.CodeRuleTimeRuleYearMonthDayHourMinute:
		return now.Format("200601021504")
	case enum.CodeRuleTimeRuleYearMonthDayHourMinuteSecond:
		return now.Format("20060102150405")
	default:
		return ""
	}
}

// buildTimeKey 构建时间键（用于区分序列）
func (uc *CodeGeneratorUseCase) buildTimeKey(generateType enum.CodeRuleGenerateType) string {
	if generateType == enum.CodeRuleGenerateTypeSerial {
		// 递增类型使用固定键
		return "SERIAL"
	}

	// 每天生成类型使用日期键
	return time.Now().In(util.TimeLocal).Format("20060102")
}

// formatSerial 格式化流水号（补齐前导零）
func (uc *CodeGeneratorUseCase) formatSerial(serial int64, length int) string {
	serialStr := strconv.FormatInt(serial, 10)

	// 如果长度不够，补齐前导零
	if len(serialStr) < length {
		serialStr = strings.Repeat("0", length-len(serialStr)) + serialStr
	}

	return serialStr
}

// calculateMaxSerial 计算给定长度的最大流水号
func (uc *CodeGeneratorUseCase) calculateMaxSerial(serialLength int32) int64 {
	// serialLength=6 时，最大值为 999999
	maxSerial := int64(1)
	for i := int32(0); i < serialLength; i++ {
		maxSerial *= 10
	}
	return maxSerial - 1
}
