package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/service"
	"github.com/xelarion/go-layout/internal/api/http/web/types"
	"github.com/xelarion/go-layout/pkg/binding"
	"github.com/xelarion/go-layout/pkg/errs"
)

// CodeRuleHandler defines the code_rule HTTP handlers.
type CodeRuleHandler struct {
	codeRuleService *service.CodeRuleService
	logger          *zap.Logger
}

// NewCodeRuleHandler creates a new instance of CodeRuleHandler.
func NewCodeRuleHandler(codeRuleService *service.CodeRuleService, logger *zap.Logger) *CodeRuleHandler {
	return &CodeRuleHandler{
		codeRuleService: codeRuleService,
		logger:          logger.Named("web_code_rule_handler"),
	}
}

// ListCodeRules godoc
//
//	@ID				ListCodeRules
//	@Summary		List CodeRules
//	@Description	Retrieves a list of CodeRules
//	@Tags			code_rule
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.ListCodeRulesReq							false	"req"
//	@Success		200	{object}	types.Response{data=types.ListCodeRulesResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/code_rules [get]
func (h *CodeRuleHandler) ListCodeRules(c *gin.Context) {
	var req types.ListCodeRulesReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.codeRuleService.ListCodeRules(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// GetCodeRuleFormData godoc
//
//	@ID				GetCodeRuleFormData
//	@Summary		Get CodeRule Form Data
//	@Description	Retrieves a single CodeRule Form Data
//	@Tags			code_rule
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer												true	"id"
//	@Param			req	query		types.GetCodeRuleFormDataReq						false	"req"
//	@Success		200	{object}	types.Response{data=types.GetCodeRuleFormDataResp}	"Success"
//	@Failure		400	{object}	types.Response										"Bad request"
//	@Failure		401	{object}	types.Response										"Unauthorized"
//	@Failure		500	{object}	types.Response										"Internal server error"
//	@Security		BearerAuth
//	@Router			/code_rules/{id}/form [get]
func (h *CodeRuleHandler) GetCodeRuleFormData(c *gin.Context) {
	var req types.GetCodeRuleFormDataReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.codeRuleService.GetCodeRuleFormData(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// UpdateCodeRule godoc
//
//	@ID				UpdateCodeRule
//	@Summary		Update CodeRule
//	@Description	Updates an existing CodeRule
//	@Tags			code_rule
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer											true	"id"
//	@Param			req	body		types.UpdateCodeRuleReq							true	"req"
//	@Success		200	{object}	types.Response{data=types.UpdateCodeRuleResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/code_rules/{id} [put]
func (h *CodeRuleHandler) UpdateCodeRule(c *gin.Context) {
	var req types.UpdateCodeRuleReq
	if err := binding.Bind(c, &req, binding.URI, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.codeRuleService.UpdateCodeRule(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("updated successfully"))
}

// GetCodeRuleTargetFormOptions godoc
//
//	@ID				GetCodeRuleTargetFormOptions
//	@Summary		Get Code Rule Target Form Options
//	@Description	Retrieves a single Code Rule Target Form Options
//	@Tags			code_rule
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.GetCodeRuleTargetFormOptionsReq						false	"req"
//	@Success		200	{object}	types.Response{data=types.GetCodeRuleTargetFormOptionsResp}	"Success"
//	@Failure		400	{object}	types.Response												"Bad request"
//	@Failure		401	{object}	types.Response												"Unauthorized"
//	@Failure		500	{object}	types.Response												"Internal server error"
//	@Security		BearerAuth
//	@Router			/code_rules/target_form/options [get]
func (h *CodeRuleHandler) GetCodeRuleTargetFormOptions(c *gin.Context) {
	var req types.GetCodeRuleTargetFormOptionsReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.codeRuleService.GetCodeRuleTargetFormOptions(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}
