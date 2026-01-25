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

// ChangeLogHandler defines the change_log HTTP handlers.
type ChangeLogHandler struct {
	changeLogService *service.ChangeLogService
	logger           *zap.Logger
}

// NewChangeLogHandler creates a new instance of ChangeLogHandler.
func NewChangeLogHandler(changeLogService *service.ChangeLogService, logger *zap.Logger) *ChangeLogHandler {
	return &ChangeLogHandler{
		changeLogService: changeLogService,
		logger:           logger.Named("web_change_log_handler"),
	}
}

// ListChangeLogs godoc
//
//	@ID				ListChangeLogs
//	@Summary		List Change Logs
//	@Description	Retrieves a list of Change Logs
//	@Tags			change_log
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.ListChangeLogsReq							false	"req"
//	@Success		200	{object}	types.Response{data=types.ListChangeLogsResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/change_logs [get]
func (h *ChangeLogHandler) ListChangeLogs(c *gin.Context) {
	var req types.ListChangeLogsReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.changeLogService.ListChangeLogs(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}
