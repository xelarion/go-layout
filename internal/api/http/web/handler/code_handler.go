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

// CodeHandler defines the code HTTP handlers.
type CodeHandler struct {
	codeService *service.CodeService
	logger      *zap.Logger
}

// NewCodeHandler creates a new instance of CodeHandler.
func NewCodeHandler(codeService *service.CodeService, logger *zap.Logger) *CodeHandler {
	return &CodeHandler{
		codeService: codeService,
		logger:      logger.Named("web_code_handler"),
	}
}

// GenerateCode godoc
//
//	@ID				GenerateCode
//	@Summary		Generate Code
//	@Description	Generate Code
//	@Tags			code
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.GenerateCodeReq						true	"req"
//	@Success		201	{object}	types.Response{data=types.GenerateCodeResp}	"Success"
//	@Failure		400	{object}	types.Response								"Bad request"
//	@Failure		401	{object}	types.Response								"Unauthorized"
//	@Failure		500	{object}	types.Response								"Internal server error"
//	@Security		BearerAuth
//	@Router			/codes/generate [post]
func (h *CodeHandler) GenerateCode(c *gin.Context) {
	var req types.GenerateCodeReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.codeService.GenerateCode(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// PreviewCode godoc
//
//	@ID				PreviewCode
//	@Summary		Preview Code
//	@Description	Preview Code
//	@Tags			code
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.PreviewCodeReq						true	"req"
//	@Success		201	{object}	types.Response{data=types.PreviewCodeResp}	"Success"
//	@Failure		400	{object}	types.Response								"Bad request"
//	@Failure		401	{object}	types.Response								"Unauthorized"
//	@Failure		500	{object}	types.Response								"Internal server error"
//	@Security		BearerAuth
//	@Router			/codes/preview [post]
func (h *CodeHandler) PreviewCode(c *gin.Context) {
	var req types.PreviewCodeReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.codeService.PreviewCode(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}
