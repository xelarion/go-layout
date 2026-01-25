package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/service"
	"github.com/xelarion/go-layout/internal/api/http/web/types"
	"github.com/xelarion/go-layout/pkg/binding"
)

// ExportHandler 导出处理器
type ExportHandler struct {
	exportService *service.ExportService
	logger        *zap.Logger
}

// NewExportHandler 创建导出处理器实例
func NewExportHandler(
	exportService *service.ExportService,
	logger *zap.Logger,
) *ExportHandler {
	return &ExportHandler{
		exportService: exportService,
		logger:        logger,
	}
}

// ListExportTasks godoc
//
//	@ID				ListExportTasks
//	@Summary		List Export Tasks
//	@Description	Retrieves a list of Export Tasks
//	@Tags			export
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.ListExportTasksReq						false	"req"
//	@Success		200	{object}	types.Response{data=types.ListExportTasksResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/export_tasks [get]
func (h *ExportHandler) ListExportTasks(c *gin.Context) {
	var req types.ListExportTasksReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(err)
		return
	}

	resp, err := h.exportService.ListExportTasks(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// GetExportFileDownloadURL godoc
//
//	@ID				GetExportFileDownloadURL
//	@Summary		Get Export File Download URL
//	@Description	Retrieves a single Export File Download URL
//	@Tags			export
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer													true	"id"
//	@Param			req	query		types.GetExportFileDownloadURLReq						false	"req"
//	@Success		200	{object}	types.Response{data=types.GetExportFileDownloadURLResp}	"Success"
//	@Failure		400	{object}	types.Response											"Bad request"
//	@Failure		401	{object}	types.Response											"Unauthorized"
//	@Failure		500	{object}	types.Response											"Internal server error"
//	@Security		BearerAuth
//	@Router			/export_tasks/{id}/download_url [get]
func (h *ExportHandler) GetExportFileDownloadURL(c *gin.Context) {
	var req types.GetExportFileDownloadURLReq
	if err := binding.Bind(c, &req, binding.URI); err != nil {
		_ = c.Error(err)
		return
	}

	resp, err := h.exportService.GetExportFileDownloadURL(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// DeleteExportTask godoc
//
//	@ID				DeleteExportTask
//	@Summary		Delete Export Task
//	@Description	Deletes an existing Export Task
//	@Tags			export
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer											true	"id"
//	@Param			req	body		types.DeleteExportTaskReq						true	"req"
//	@Success		204	{object}	types.Response{data=types.DeleteExportTaskResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/export_tasks/{id} [delete]
func (h *ExportHandler) DeleteExportTask(c *gin.Context) {
	var req types.DeleteExportTaskReq
	if err := binding.Bind(c, &req, binding.URI); err != nil {
		_ = c.Error(err)
		return
	}

	resp, err := h.exportService.DeleteExportTask(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("deleted successfully"))
}
