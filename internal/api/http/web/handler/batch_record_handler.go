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

// BatchRecordHandler defines the batch record HTTP handlers.
type BatchRecordHandler struct {
	batchRecordService *service.BatchRecordService
	logger             *zap.Logger
}

// NewBatchRecordHandler creates a new instance of BatchRecordHandler.
func NewBatchRecordHandler(batchRecordService *service.BatchRecordService, logger *zap.Logger) *BatchRecordHandler {
	return &BatchRecordHandler{
		batchRecordService: batchRecordService,
		logger:             logger.Named("web_batch_record_handler"),
	}
}

// ListBatchRecords godoc
//
//	@ID				ListBatchRecords
//	@Summary		List Batch Records
//	@Description	Retrieves a list of Batch Records
//	@Tags			batch_record
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.ListBatchRecordsReq						false	"req"
//	@Success		200	{object}	types.Response{data=types.ListBatchRecordsResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/batch_records [get]
func (h *BatchRecordHandler) ListBatchRecords(c *gin.Context) {
	var req types.ListBatchRecordsReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.batchRecordService.ListBatchRecords(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// GetBatchRecord godoc
//
//	@ID				GetBatchRecord
//	@Summary		Get Batch Record
//	@Description	Retrieves a single Batch Record
//	@Tags			batch_record
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer											true	"id"
//	@Param			req	query		types.GetBatchRecordReq							false	"req"
//	@Success		200	{object}	types.Response{data=types.GetBatchRecordResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/batch_records/{id} [get]
func (h *BatchRecordHandler) GetBatchRecord(c *gin.Context) {
	var req types.GetBatchRecordReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.batchRecordService.GetBatchRecord(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// GetBatchRecordOperationTypeOptions godoc
//
//	@ID				GetBatchRecordOperationTypeOptions
//	@Summary		Get Batch Record Operation Type Options
//	@Description	Retrieves a single Batch Record Operation Type Options
//	@Tags			batch_record
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.GetBatchRecordOperationTypeOptionsReq							false	"req"
//	@Success		200	{object}	types.Response{data=types.GetBatchRecordOperationTypeOptionsResp}	"Success"
//	@Failure		400	{object}	types.Response														"Bad request"
//	@Failure		401	{object}	types.Response														"Unauthorized"
//	@Failure		500	{object}	types.Response														"Internal server error"
//	@Security		BearerAuth
//	@Router			/batch_records/operation_type/options [get]
func (h *BatchRecordHandler) GetBatchRecordOperationTypeOptions(c *gin.Context) {
	var req types.GetBatchRecordOperationTypeOptionsReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.batchRecordService.GetBatchRecordOperationTypeOptions(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}
