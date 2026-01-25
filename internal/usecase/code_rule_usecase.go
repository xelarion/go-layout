// Package usecase contains business logic.
package usecase

import (
	"context"

	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/pkg/errs"
)

// CodeRule contains code_rule data
type CodeRule struct {
	*model.CodeRule
}

// UpdateCodeRuleParams contains all parameters needed to update a code_rule
type UpdateCodeRuleParams struct {
	UpdateRecordCommonParams
	ID            uint
	Prefix        string
	TimeRule      enum.CodeRuleTimeRule
	GenerateType  enum.CodeRuleGenerateType
	InitialSerial int32
	SerialLength  int32
	Step          uint

	// Fields to track which values are explicitly set
	PrefixSet        bool
	TimeRuleSet      bool
	GenerateTypeSet  bool
	InitialSerialSet bool
	SerialLengthSet  bool
	StepSet          bool
}

// CodeRuleRepository defines methods for code_rule data access.
type CodeRuleRepository interface {
	Create(ctx context.Context, codeRule *model.CodeRule) error
	List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*model.CodeRule, int, error)
	IsExists(ctx context.Context, filters map[string]any, notFilters map[string]any) (bool, error)
	Count(ctx context.Context, filters map[string]any, notFilters map[string]any) (int64, error)
	FindByID(ctx context.Context, id uint) (*model.CodeRule, error)
	FindBySlug(ctx context.Context, slug string) (*model.CodeRule, error)
	Update(ctx context.Context, codeRule *model.CodeRule, params map[string]any) error
	Delete(ctx context.Context, id uint) error
}

// CodeRuleUseCase implements business logic for code_rule operations.
type CodeRuleUseCase struct {
	tx           Transaction
	codeRuleRepo CodeRuleRepository
	userRepo     UserRepository
}

// NewCodeRuleUseCase creates a new instance of CodeRuleUseCase.
func NewCodeRuleUseCase(tx Transaction, repo CodeRuleRepository, userRepo UserRepository) *CodeRuleUseCase {
	return &CodeRuleUseCase{
		tx:           tx,
		codeRuleRepo: repo,
		userRepo:     userRepo,
	}
}

// List returns a list of code_rules with pagination and filtering.
func (uc *CodeRuleUseCase) List(ctx context.Context, filters map[string]any, limit, offset int, sortClause string) ([]*CodeRule, int, error) {
	codeRules, count, err := uc.codeRuleRepo.List(ctx, filters, limit, offset, sortClause)
	if err != nil {
		return nil, 0, err
	}

	records := make([]*CodeRule, 0, len(codeRules))
	for _, codeRule := range codeRules {
		records = append(records, &CodeRule{
			CodeRule: codeRule,
		})
	}

	return records, count, nil
}

// GetByID retrieves a code_rule by ID.
func (uc *CodeRuleUseCase) GetByID(ctx context.Context, id uint) (*CodeRule, error) {
	codeRule, err := uc.codeRuleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &CodeRule{
		CodeRule: codeRule,
	}, nil
}

// Update updates an existing code_rule.
func (uc *CodeRuleUseCase) Update(ctx context.Context, params UpdateCodeRuleParams) error {
	// Get the existing code_rule
	codeRule, err := uc.codeRuleRepo.FindByID(ctx, params.ID)
	if err != nil {
		return err
	}

	newSerialLength := codeRule.SerialLength
	newInitialSerial := codeRule.InitialSerial

	updates := map[string]any{}
	if params.PrefixSet {
		updates["prefix"] = params.Prefix
	}

	if params.TimeRuleSet {
		updates["time_rule"] = params.TimeRule
	}

	if params.GenerateTypeSet {
		updates["generate_type"] = params.GenerateType
	}

	if params.InitialSerialSet {
		newInitialSerial = params.InitialSerial
		updates["initial_serial"] = params.InitialSerial
	}

	if params.SerialLengthSet {
		newSerialLength = params.SerialLength
		updates["serial_length"] = params.SerialLength
	}

	maxSerial := uc.calculateMaxSerial(newSerialLength)
	if int64(newInitialSerial) > maxSerial {
		return errs.NewBusiness("初始流水号不能大于流水号长度支持的最大值")
	}

	if len(updates) > 0 {
		updates["updated_by"] = params.UpdatedBy
		err = uc.tx.Transaction(ctx, func(ctx context.Context) error {
			if err := uc.codeRuleRepo.Update(ctx, codeRule, updates); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// GetCodeRuleOptions 获取编号规则作为选项数据
func (uc *CodeRuleUseCase) GetCodeRuleOptions(ctx context.Context, filters map[string]any, sortClause string) ([]*CodeRule, error) {
	codeRules, _, err := uc.codeRuleRepo.List(ctx, filters, -1, -1, sortClause)
	if err != nil {
		return nil, err
	}

	records := make([]*CodeRule, 0, len(codeRules))
	for _, codeRule := range codeRules {
		record := &CodeRule{
			CodeRule: codeRule,
		}
		records = append(records, record)
	}

	return records, nil
}

// calculateMaxSerial 计算给定长度的最大流水号
func (uc *CodeRuleUseCase) calculateMaxSerial(serialLength int32) int64 {
	// serialLength=6 时，最大值为 999999
	maxSerial := int64(1)
	for i := int32(0); i < serialLength; i++ {
		maxSerial *= 10
	}
	return maxSerial - 1
}
