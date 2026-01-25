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

// BatchDetailHandler defines the batch detail HTTP handlers.
type BatchDetailHandler struct {
	batchDetailService *service.BatchDetailService
	logger             *zap.Logger
}

// NewBatchDetailHandler creates a new instance of BatchDetailHandler.
func NewBatchDetailHandler(batchDetailService *service.BatchDetailService, logger *zap.Logger) *BatchDetailHandler {
	return &BatchDetailHandler{
		batchDetailService: batchDetailService,
		logger:             logger.Named("web_batch_detail_handler"),
	}
}

// ListBatchDetails godoc
//
//	@ID				ListBatchDetails
//	@Summary		List Batch Details
//	@Description	Retrieves a list of Batch Details
//	@Tags			batch_detail
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.ListBatchDetailsReq						false	"req"
//	@Success		200	{object}	types.Response{data=types.ListBatchDetailsResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/batch_details [get]
func (h *BatchDetailHandler) ListBatchDetails(c *gin.Context) {
	var req types.ListBatchDetailsReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.batchDetailService.ListBatchDetails(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}
