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

// NotificationHandler defines the notification HTTP handlers.
type NotificationHandler struct {
	notificationService *service.NotificationService
	logger              *zap.Logger
}

// NewNotificationHandler creates a new instance of NotificationHandler.
func NewNotificationHandler(notificationService *service.NotificationService, logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		logger:              logger.Named("web_notification_handler"),
	}
}

// ListNotifications godoc
//
//	@ID				ListNotifications
//	@Summary		List Notifications
//	@Description	Retrieves a list of Notifications
//	@Tags			notification
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.ListNotificationsReq							false	"req"
//	@Success		200	{object}	types.Response{data=types.ListNotificationsResp}	"Success"
//	@Failure		400	{object}	types.Response										"Bad request"
//	@Failure		401	{object}	types.Response										"Unauthorized"
//	@Failure		500	{object}	types.Response										"Internal server error"
//	@Security		BearerAuth
//	@Router			/notifications [get]
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	var req types.ListNotificationsReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.notificationService.ListNotifications(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// GetRecentNotifications godoc
//
//	@ID				GetRecentNotifications
//	@Summary		Get Recent Notifications
//	@Description	Retrieves a single Recent Notifications
//	@Tags			notification
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.GetRecentNotificationsReq							false	"req"
//	@Success		200	{object}	types.Response{data=types.GetRecentNotificationsResp}	"Success"
//	@Failure		400	{object}	types.Response											"Bad request"
//	@Failure		401	{object}	types.Response											"Unauthorized"
//	@Failure		500	{object}	types.Response											"Internal server error"
//	@Security		BearerAuth
//	@Router			/notifications/recent [get]
func (h *NotificationHandler) GetRecentNotifications(c *gin.Context) {
	var req types.GetRecentNotificationsReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.notificationService.GetRecentNotifications(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// GetNotification godoc
//
//	@ID				GetNotification
//	@Summary		Get Notification
//	@Description	Retrieves a single Notification
//	@Tags			notification
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer											true	"id"
//	@Param			req	query		types.GetNotificationReq						false	"req"
//	@Success		200	{object}	types.Response{data=types.GetNotificationResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/notifications/{id} [get]
func (h *NotificationHandler) GetNotification(c *gin.Context) {
	var req types.GetNotificationReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.notificationService.GetNotification(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// DeleteNotification godoc
//
//	@ID				DeleteNotification
//	@Summary		Delete Notification
//	@Description	Deletes an existing Notification
//	@Tags			notification
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer												true	"id"
//	@Param			req	body		types.DeleteNotificationReq							true	"req"
//	@Success		204	{object}	types.Response{data=types.DeleteNotificationResp}	"Success"
//	@Failure		400	{object}	types.Response										"Bad request"
//	@Failure		401	{object}	types.Response										"Unauthorized"
//	@Failure		500	{object}	types.Response										"Internal server error"
//	@Security		BearerAuth
//	@Router			/notifications/{id} [delete]
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	var req types.DeleteNotificationReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.notificationService.DeleteNotification(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("删除成功"))
}

// BatchDeleteNotifications godoc
//
//	@ID				BatchDeleteNotifications
//	@Summary		Batch Delete Notifications
//	@Description	Batch Delete Notifications
//	@Tags			notification
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.BatchDeleteNotificationsReq						true	"req"
//	@Success		200	{object}	types.Response{data=types.BatchDeleteNotificationsResp}	"Success"
//	@Failure		400	{object}	types.Response											"Bad request"
//	@Failure		401	{object}	types.Response											"Unauthorized"
//	@Failure		500	{object}	types.Response											"Internal server error"
//	@Security		BearerAuth
//	@Router			/notifications/batch-delete [post]
func (h *NotificationHandler) BatchDeleteNotifications(c *gin.Context) {
	var req types.BatchDeleteNotificationsReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.notificationService.BatchDeleteNotifications(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// BatchReadNotifications godoc
//
//	@ID				BatchReadNotifications
//	@Summary		Batch Read Notifications
//	@Description	Batch Read Notifications
//	@Tags			notification
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.BatchReadNotificationsReq							true	"req"
//	@Success		200	{object}	types.Response{data=types.BatchReadNotificationsResp}	"Success"
//	@Failure		400	{object}	types.Response											"Bad request"
//	@Failure		401	{object}	types.Response											"Unauthorized"
//	@Failure		500	{object}	types.Response											"Internal server error"
//	@Security		BearerAuth
//	@Router			/notifications/batch-read [post]
func (h *NotificationHandler) BatchReadNotifications(c *gin.Context) {
	var req types.BatchReadNotificationsReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.notificationService.BatchReadNotifications(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// ReadAllNotifications godoc
//
//	@ID				ReadAllNotifications
//	@Summary		Read All Notifications
//	@Description	Read All Notifications
//	@Tags			notification
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.ReadAllNotificationsReq						true	"req"
//	@Success		200	{object}	types.Response{data=types.ReadAllNotificationsResp}	"Success"
//	@Failure		400	{object}	types.Response										"Bad request"
//	@Failure		401	{object}	types.Response										"Unauthorized"
//	@Failure		500	{object}	types.Response										"Internal server error"
//	@Security		BearerAuth
//	@Router			/notifications/read-all [post]
func (h *NotificationHandler) ReadAllNotifications(c *gin.Context) {
	var req types.ReadAllNotificationsReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.notificationService.ReadAllNotifications(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// GetTodoStats godoc
//
//	@ID				GetTodoStats
//	@Summary		Get Todo Stats
//	@Description	Retrieves a single Todo Stats
//	@Tags			notification
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.GetTodoStatsReq						false	"req"
//	@Success		200	{object}	types.Response{data=types.GetTodoStatsResp}	"Success"
//	@Failure		400	{object}	types.Response								"Bad request"
//	@Failure		401	{object}	types.Response								"Unauthorized"
//	@Failure		500	{object}	types.Response								"Internal server error"
//	@Security		BearerAuth
//	@Router			/todo_stats [get]
func (h *NotificationHandler) GetTodoStats(c *gin.Context) {
	var req types.GetTodoStatsReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.notificationService.GetTodoStats(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}
