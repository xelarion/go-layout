package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	appctx "github.com/xelarion/go-layout/internal/context"
	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/model"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/errs"
)

// CodeRuleService handles code_rule-related services.
type CodeRuleService struct {
	*BaseService         // 嵌入 BaseService 提供数据权限检查方法
	codeGeneratorUseCase *usecase.CodeGeneratorUseCase
	codeRuleUseCase      *usecase.CodeRuleUseCase
	logger               *zap.Logger
}

// NewCodeRuleService creates a new CodeRuleService.
func NewCodeRuleService(
	baseService *BaseService,
	codeGeneratorUseCase *usecase.CodeGeneratorUseCase,
	codeRuleUseCase *usecase.CodeRuleUseCase,
	logger *zap.Logger,
) *CodeRuleService {
	return &CodeRuleService{
		BaseService:          baseService,
		codeGeneratorUseCase: codeGeneratorUseCase,
		codeRuleUseCase:      codeRuleUseCase,
		logger:               logger,
	}
}

// ListCodeRules lists code_rules with pagination.
func (s *CodeRuleService) ListCodeRules(ctx context.Context, req *types.ListCodeRulesReq) (*types.ListCodeRulesResp, error) {
	filters := map[string]any{}
	if req.Key != "" {
		filters["key"] = req.Key
	}

	if req.GenerateType != "" {
		generateType, err := enum.GetCodeRuleGenerateType(req.GenerateType)
		if err != nil {
			return nil, err
		}
		filters["generate_type"] = generateType
	}

	if req.TargetForm != "" {
		filters["target_form"] = req.TargetForm
	}

	// 应用数据权限过滤（传入空的 CodeRule 实例用于获取字段映射）
	if err := s.ApplyDataScope(ctx, filters, &model.CodeRule{}); err != nil {
		return nil, err
	}

	codeRules, count, err := s.codeRuleUseCase.List(ctx, filters, req.GetLimit(), req.GetOffset(), req.GetSortClause())
	if err != nil {
		return nil, err
	}

	respResults := make([]types.ListCodeRulesRespResult, 0, len(codeRules))
	for _, codeRule := range codeRules {
		firstCode, endCode, err := s.codeGeneratorUseCase.PreviewCode(ctx, codeRule.Prefix, codeRule.TimeRule, codeRule.GenerateType, codeRule.InitialSerial,
			codeRule.SerialLength, codeRule.Step)
		if err != nil {
			return nil, err
		}
		u := types.ListCodeRulesRespResult{
			ID:            codeRule.ID,
			TargetForm:    codeRule.TargetForm,
			Name:          codeRule.Name,
			Prefix:        codeRule.Prefix,
			TimeRule:      enum.CodeRuleTimeRuleText[codeRule.TimeRule],
			GenerateType:  enum.CodeRuleGenerateTypeText[codeRule.GenerateType],
			InitialSerial: codeRule.InitialSerial,
			SerialLength:  codeRule.SerialLength,
			Step:          codeRule.Step,
			Preview:       firstCode,
			GenerateRule:  fmt.Sprintf("%s~%s", firstCode, endCode),
		}
		respResults = append(respResults, u)
	}

	return &types.ListCodeRulesResp{
		Results:    respResults,
		Pagination: types.NewPageResp(count, req.GetPage(), req.GetPageSize()),
	}, nil
}

// GetCodeRuleFormData provides data needed for code_rule forms (update).
func (s *CodeRuleService) GetCodeRuleFormData(ctx context.Context, req *types.GetCodeRuleFormDataReq) (*types.GetCodeRuleFormDataResp, error) {
	codeRule, err := s.codeRuleUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 检查数据权限
	if err := s.CheckRecordAccess(ctx, codeRule.CodeRule); err != nil {
		return nil, err
	}

	firstCode, endCode, err := s.codeGeneratorUseCase.PreviewCode(ctx,
		codeRule.Prefix,
		codeRule.TimeRule,
		codeRule.GenerateType,
		codeRule.InitialSerial,
		codeRule.SerialLength,
		codeRule.Step,
	)
	if err != nil {
		return nil, err
	}

	return &types.GetCodeRuleFormDataResp{
		ID:            codeRule.ID,
		TargetForm:    codeRule.TargetForm,
		Name:          codeRule.Name,
		Prefix:        codeRule.Prefix,
		TimeRule:      enum.CodeRuleTimeRuleSlug[codeRule.TimeRule],
		GenerateType:  enum.CodeRuleGenerateTypeSlug[codeRule.GenerateType],
		InitialSerial: codeRule.InitialSerial,
		SerialLength:  codeRule.SerialLength,
		Step:          codeRule.Step,
		Preview:       firstCode,
		GenerateRule:  fmt.Sprintf("%s~%s", firstCode, endCode),
	}, nil
}

// UpdateCodeRule updates a code_rule.
func (s *CodeRuleService) UpdateCodeRule(ctx context.Context, req *types.UpdateCodeRuleReq) (*types.UpdateCodeRuleResp, error) {
	current := appctx.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("unauthorized").WithReason(errs.ReasonUnauthorized)
	}

	// 先获取编码规则，用于权限检查
	codeRule, err := s.codeRuleUseCase.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 检查数据权限
	if err := s.CheckRecordAccess(ctx, codeRule.CodeRule); err != nil {
		return nil, err
	}

	// Create update params
	params := usecase.UpdateCodeRuleParams{
		UpdateRecordCommonParams: usecase.UpdateRecordCommonParams{
			UpdatedBy: current.UserID,
		},
		ID: req.ID,
	}

	if req.Prefix != "" {
		params.Prefix = req.Prefix
		params.PrefixSet = true
	}

	if req.TimeRule != "" {
		timeRule, err := enum.GetCodeRuleTimeRule(req.TimeRule)
		if err != nil {
			return nil, err
		}
		params.TimeRule = timeRule
		params.TimeRuleSet = true
	}

	if req.GenerateType != "" {
		generateType, err := enum.GetCodeRuleGenerateType(req.GenerateType)
		if err != nil {
			return nil, err
		}
		params.GenerateType = generateType
		params.GenerateTypeSet = true
	}

	if req.InitialSerial != 0 {
		params.InitialSerial = req.InitialSerial
		params.InitialSerialSet = true
	}

	if req.SerialLength != 0 {
		params.SerialLength = req.SerialLength
		params.SerialLengthSet = true
	}

	if err := s.codeRuleUseCase.Update(ctx, params); err != nil {
		return nil, err
	}

	return &types.UpdateCodeRuleResp{}, nil
}

func (s *CodeRuleService) GetCodeRuleTargetFormOptions(ctx context.Context, req *types.GetCodeRuleTargetFormOptionsReq) (*types.GetCodeRuleTargetFormOptionsResp, error) {
	filters := map[string]any{}

	codeRules, err := s.codeRuleUseCase.GetCodeRuleOptions(ctx, filters, req.GetSortClause())
	if err != nil {
		return nil, err
	}

	results := make([]types.GetCodeRuleTargetFormOptionsRespResult, 0, len(codeRules))
	for _, codeRule := range codeRules {
		results = append(results, types.GetCodeRuleTargetFormOptionsRespResult{
			GetOptionsRespResult: types.GetOptionsRespResult{
				Label: codeRule.TargetForm,
				Value: codeRule.TargetForm,
			},
		})
	}

	return &types.GetCodeRuleTargetFormOptionsResp{
		Results: results,
	}, nil
}
