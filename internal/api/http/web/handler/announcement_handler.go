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

// AnnouncementHandler defines the announcement HTTP handlers.
type AnnouncementHandler struct {
	announcementService *service.AnnouncementService
	logger              *zap.Logger
}

// NewAnnouncementHandler creates a new instance of AnnouncementHandler.
func NewAnnouncementHandler(announcementService *service.AnnouncementService, logger *zap.Logger) *AnnouncementHandler {
	return &AnnouncementHandler{
		announcementService: announcementService,
		logger:              logger.Named("web_announcement_handler"),
	}
}

// CreateAnnouncement godoc
//
//	@ID				CreateAnnouncement
//	@Summary		Create Announcement
//	@Description	Creates a new Announcement
//	@Tags			announcement
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.CreateAnnouncementReq							true	"req"
//	@Success		201	{object}	types.Response{data=types.CreateAnnouncementResp}	"Success"
//	@Failure		400	{object}	types.Response										"Bad request"
//	@Failure		401	{object}	types.Response										"Unauthorized"
//	@Failure		500	{object}	types.Response										"Internal server error"
//	@Security		BearerAuth
//	@Router			/announcements [post]
func (h *AnnouncementHandler) CreateAnnouncement(c *gin.Context) {
	var req types.CreateAnnouncementReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.announcementService.CreateAnnouncement(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, types.Success(resp).WithMessage("操作成功"))
}

// ListAnnouncements godoc
//
//	@ID				ListAnnouncements
//	@Summary		List Announcements
//	@Description	Retrieves a list of Announcements
//	@Tags			announcement
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.ListAnnouncementsReq							false	"req"
//	@Success		200	{object}	types.Response{data=types.ListAnnouncementsResp}	"Success"
//	@Failure		400	{object}	types.Response										"Bad request"
//	@Failure		401	{object}	types.Response										"Unauthorized"
//	@Failure		500	{object}	types.Response										"Internal server error"
//	@Security		BearerAuth
//	@Router			/announcements [get]
func (h *AnnouncementHandler) ListAnnouncements(c *gin.Context) {
	var req types.ListAnnouncementsReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.announcementService.ListAnnouncements(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// GetAnnouncement godoc
//
//	@ID				GetAnnouncement
//	@Summary		Get Announcement
//	@Description	Retrieves a single Announcement
//	@Tags			announcement
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer											true	"id"
//	@Param			req	query		types.GetAnnouncementReq						false	"req"
//	@Success		200	{object}	types.Response{data=types.GetAnnouncementResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/announcements/{id} [get]
func (h *AnnouncementHandler) GetAnnouncement(c *gin.Context) {
	var req types.GetAnnouncementReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.announcementService.GetAnnouncement(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// GetAnnouncementFormData godoc
//
//	@ID				GetAnnouncementFormData
//	@Summary		Get Announcement Form Data
//	@Description	Retrieves a single Announcement Form Data
//	@Tags			announcement
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer													true	"id"
//	@Param			req	query		types.GetAnnouncementFormDataReq						false	"req"
//	@Success		200	{object}	types.Response{data=types.GetAnnouncementFormDataResp}	"Success"
//	@Failure		400	{object}	types.Response											"Bad request"
//	@Failure		401	{object}	types.Response											"Unauthorized"
//	@Failure		500	{object}	types.Response											"Internal server error"
//	@Security		BearerAuth
//	@Router			/announcements/{id}/form [get]
func (h *AnnouncementHandler) GetAnnouncementFormData(c *gin.Context) {
	var req types.GetAnnouncementFormDataReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.announcementService.GetAnnouncementFormData(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// UpdateAnnouncement godoc
//
//	@ID				UpdateAnnouncement
//	@Summary		Update Announcement
//	@Description	Updates an existing Announcement
//	@Tags			announcement
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer												true	"id"
//	@Param			req	body		types.UpdateAnnouncementReq							true	"req"
//	@Success		200	{object}	types.Response{data=types.UpdateAnnouncementResp}	"Success"
//	@Failure		400	{object}	types.Response										"Bad request"
//	@Failure		401	{object}	types.Response										"Unauthorized"
//	@Failure		500	{object}	types.Response										"Internal server error"
//	@Security		BearerAuth
//	@Router			/announcements/{id} [put]
func (h *AnnouncementHandler) UpdateAnnouncement(c *gin.Context) {
	var req types.UpdateAnnouncementReq
	if err := binding.Bind(c, &req, binding.URI, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.announcementService.UpdateAnnouncement(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("操作成功"))
}

// DeleteAnnouncement godoc
//
//	@ID				DeleteAnnouncement
//	@Summary		Delete Announcement
//	@Description	Deletes an existing Announcement
//	@Tags			announcement
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer												true	"id"
//	@Param			req	body		types.DeleteAnnouncementReq							true	"req"
//	@Success		204	{object}	types.Response{data=types.DeleteAnnouncementResp}	"Success"
//	@Failure		400	{object}	types.Response										"Bad request"
//	@Failure		401	{object}	types.Response										"Unauthorized"
//	@Failure		500	{object}	types.Response										"Internal server error"
//	@Security		BearerAuth
//	@Router			/announcements/{id} [delete]
func (h *AnnouncementHandler) DeleteAnnouncement(c *gin.Context) {
	var req types.DeleteAnnouncementReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.announcementService.DeleteAnnouncement(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("删除成功"))
}

// BatchDeleteAnnouncements godoc
//
//	@ID				BatchDeleteAnnouncements
//	@Summary		Batch Delete Announcements
//	@Description	Batch Delete Announcements
//	@Tags			announcement
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.BatchDeleteAnnouncementsReq						true	"req"
//	@Success		200	{object}	types.Response{data=types.BatchDeleteAnnouncementsResp}	"Success"
//	@Failure		400	{object}	types.Response											"Bad request"
//	@Failure		401	{object}	types.Response											"Unauthorized"
//	@Failure		500	{object}	types.Response											"Internal server error"
//	@Security		BearerAuth
//	@Router			/announcements/batch-delete [post]
func (h *AnnouncementHandler) BatchDeleteAnnouncements(c *gin.Context) {
	var req types.BatchDeleteAnnouncementsReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.announcementService.BatchDeleteAnnouncements(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// GetAnnouncementPublishersOptions godoc
//
//	@ID				GetAnnouncementPublishersOptions
//	@Summary		Get Announcement Publishers Options
//	@Description	Retrieves a single Announcement Publishers Options
//	@Tags			announcement
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.GetAnnouncementPublishersOptionsReq						false	"req"
//	@Success		200	{object}	types.Response{data=types.GetAnnouncementPublishersOptionsResp}	"Success"
//	@Failure		400	{object}	types.Response													"Bad request"
//	@Failure		401	{object}	types.Response													"Unauthorized"
//	@Failure		500	{object}	types.Response													"Internal server error"
//	@Security		BearerAuth
//	@Router			/announcements/publishers/options [get]
func (h *AnnouncementHandler) GetAnnouncementPublishersOptions(c *gin.Context) {
	var req types.GetAnnouncementPublishersOptionsReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.announcementService.GetAnnouncementPublishersOptions(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}
