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

// PositionHandler defines the position HTTP handlers.
type PositionHandler struct {
	positionService *service.PositionService
	logger          *zap.Logger
}

// NewPositionHandler creates a new instance of PositionHandler.
func NewPositionHandler(positionService *service.PositionService, logger *zap.Logger) *PositionHandler {
	return &PositionHandler{
		positionService: positionService,
		logger:          logger.Named("web_position_handler"),
	}
}

// CreatePosition godoc
//
//	@ID				CreatePosition
//	@Summary		Create Position
//	@Description	Creates a new Position
//	@Tags			position
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.CreatePositionReq							true	"req"
//	@Success		201	{object}	types.Response{data=types.CreatePositionResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/positions [post]
func (h *PositionHandler) CreatePosition(c *gin.Context) {
	var req types.CreatePositionReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.positionService.CreatePosition(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, types.Success(resp).WithMessage("created successfully"))
}

// ListPositions godoc
//
//	@ID				ListPositions
//	@Summary		List Positions
//	@Description	Retrieves a list of Positions
//	@Tags			position
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.ListPositionsReq							false	"req"
//	@Success		200	{object}	types.Response{data=types.ListPositionsResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/positions [get]
func (h *PositionHandler) ListPositions(c *gin.Context) {
	var req types.ListPositionsReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.positionService.ListPositions(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// GetPositionFormData godoc
//
//	@ID				GetPositionFormData
//	@Summary		Get Position Form Data
//	@Description	Retrieves a single Position Form Data
//	@Tags			position
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer												true	"id"
//	@Param			req	query		types.GetPositionFormDataReq						false	"req"
//	@Success		200	{object}	types.Response{data=types.GetPositionFormDataResp}	"Success"
//	@Failure		400	{object}	types.Response										"Bad request"
//	@Failure		401	{object}	types.Response										"Unauthorized"
//	@Failure		500	{object}	types.Response										"Internal server error"
//	@Security		BearerAuth
//	@Router			/positions/{id}/form [get]
func (h *PositionHandler) GetPositionFormData(c *gin.Context) {
	var req types.GetPositionFormDataReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.positionService.GetPositionFormData(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// UpdatePosition godoc
//
//	@ID				UpdatePosition
//	@Summary		Update Position
//	@Description	Updates an existing Position
//	@Tags			position
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer											true	"id"
//	@Param			req	body		types.UpdatePositionReq							true	"req"
//	@Success		200	{object}	types.Response{data=types.UpdatePositionResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/positions/{id} [put]
func (h *PositionHandler) UpdatePosition(c *gin.Context) {
	var req types.UpdatePositionReq
	if err := binding.Bind(c, &req, binding.URI, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.positionService.UpdatePosition(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("updated successfully"))
}

// DeletePosition godoc
//
//	@ID				DeletePosition
//	@Summary		Delete Position
//	@Description	Deletes an existing Position
//	@Tags			position
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer											true	"id"
//	@Param			req	body		types.DeletePositionReq							true	"req"
//	@Success		204	{object}	types.Response{data=types.DeletePositionResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/positions/{id} [delete]
func (h *PositionHandler) DeletePosition(c *gin.Context) {
	var req types.DeletePositionReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.positionService.DeletePosition(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("deleted successfully"))
}

// GetPositionOptions godoc
//
//	@ID				GetPositionOptions
//	@Summary		Get Position Options
//	@Description	Retrieves a single Position Options
//	@Tags			position
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.GetPositionOptionsReq							false	"req"
//	@Success		200	{object}	types.Response{data=types.GetPositionOptionsResp}	"Success"
//	@Failure		400	{object}	types.Response										"Bad request"
//	@Failure		401	{object}	types.Response										"Unauthorized"
//	@Failure		500	{object}	types.Response										"Internal server error"
//	@Security		BearerAuth
//	@Router			/positions/options [get]
func (h *PositionHandler) GetPositionOptions(c *gin.Context) {
	var req types.GetPositionOptionsReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.positionService.GetPositionOptions(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// BatchDeletePositions godoc
//
//	@ID				BatchDeletePositions
//	@Summary		Batch Delete Positions
//	@Description	Batch Delete Positions
//	@Tags			position
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.BatchDeletePositionsReq						true	"req"
//	@Success		201	{object}	types.Response{data=types.BatchDeletePositionsResp}	"Success"
//	@Failure		400	{object}	types.Response										"Bad request"
//	@Failure		401	{object}	types.Response										"Unauthorized"
//	@Failure		500	{object}	types.Response										"Internal server error"
//	@Security		BearerAuth
//	@Router			/positions/batch-delete [post]
func (h *PositionHandler) BatchDeletePositions(c *gin.Context) {
	var req types.BatchDeletePositionsReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.positionService.BatchDeletePositions(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}
