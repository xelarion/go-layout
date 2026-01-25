package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	"github.com/xelarion/go-layout/internal/enum"
	"github.com/xelarion/go-layout/internal/usecase"
)

// CodeService handles code-related services.
type CodeService struct {
	codeGeneratorUseCase *usecase.CodeGeneratorUseCase
	logger               *zap.Logger
}

// NewCodeService creates a new CodeService.
func NewCodeService(codeGeneratorUseCase *usecase.CodeGeneratorUseCase, logger *zap.Logger) *CodeService {
	return &CodeService{
		codeGeneratorUseCase: codeGeneratorUseCase,
		logger:               logger,
	}
}

// GenerateCode generates a new code.
func (s *CodeService) GenerateCode(ctx context.Context, req *types.GenerateCodeReq) (*types.GenerateCodeResp, error) {
	code, err := s.codeGeneratorUseCase.GenerateCode(ctx, req.RuleSlug)
	if err != nil {
		return nil, err
	}

	return &types.GenerateCodeResp{
		Code: code,
	}, nil
}

// PreviewCode previews a code without consuming sequence number.
func (s *CodeService) PreviewCode(ctx context.Context, req *types.PreviewCodeReq) (*types.PreviewCodeResp, error) {
	timeRule, err := enum.GetCodeRuleTimeRule(req.TimeRule)
	if err != nil {
		return nil, err
	}

	generateType, err := enum.GetCodeRuleGenerateType(req.GenerateType)
	if err != nil {
		return nil, err
	}

	firstCode, endCode, err := s.codeGeneratorUseCase.PreviewCode(ctx, req.Prefix, timeRule, generateType, req.InitialSerial,
		req.SerialLength, req.Step)
	if err != nil {
		return nil, err
	}

	return &types.PreviewCodeResp{
		Code:         firstCode,
		GenerateRule: fmt.Sprintf("%s~%s", firstCode, endCode),
	}, nil
}
