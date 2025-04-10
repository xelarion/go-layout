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

// UserHandler defines the user HTTP handlers.
type UserHandler struct {
	userService *service.UserService
	logger      *zap.Logger
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(userService *service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger.Named("web_user_handler"),
	}
}

// CreateUser godoc
//	@Summary		Create User
//	@Description	Creates a new User
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.CreateUserReq							true	"req"
//	@Success		201	{object}	types.Response{data=types.CreateUserResp}	"Success"
//	@Failure		400	{object}	types.Response								"Bad request"
//	@Failure		401	{object}	types.Response								"Unauthorized"
//	@Failure		500	{object}	types.Response								"Internal server error"
//	@Security		BearerAuth
//	@Router			/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req types.CreateUserReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, types.Success(resp).WithMessage("User created successfully"))
}

// ListUsers godoc
//	@Summary		List Users
//	@Description	Retrieves a list of Users
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			req	query		types.ListUsersReq							false	"req"
//	@Success		200	{object}	types.Response{data=types.ListUsersResp}	"Success"
//	@Failure		400	{object}	types.Response								"Bad request"
//	@Failure		401	{object}	types.Response								"Unauthorized"
//	@Failure		500	{object}	types.Response								"Internal server error"
//	@Security		BearerAuth
//	@Router			/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var req types.ListUsersReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.userService.ListUsers(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// GetUser godoc
//	@Summary		Get User
//	@Description	Retrieves a single User
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer									true	"id"
//	@Param			req	query		types.GetUserReq						false	"req"
//	@Success		200	{object}	types.Response{data=types.GetUserResp}	"Success"
//	@Failure		400	{object}	types.Response							"Bad request"
//	@Failure		401	{object}	types.Response							"Unauthorized"
//	@Failure		500	{object}	types.Response							"Internal server error"
//	@Security		BearerAuth
//	@Router			/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	var req types.GetUserReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.userService.GetUser(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if resp == nil {
		c.JSON(http.StatusNotFound, types.Error(types.CodeNotFound, "User not found"))
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// GetUserFormData godoc
//	@Summary		Get User Form Data
//	@Description	Retrieves a single User Form Data
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer											true	"id"
//	@Param			req	query		types.GetUserFormDataReq						false	"req"
//	@Success		200	{object}	types.Response{data=types.GetUserFormDataResp}	"Success"
//	@Failure		400	{object}	types.Response									"Bad request"
//	@Failure		401	{object}	types.Response									"Unauthorized"
//	@Failure		500	{object}	types.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/users/{id}/form [get]
func (h *UserHandler) GetUserFormData(c *gin.Context) {
	var req types.GetUserFormDataReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.userService.GetUserFormData(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// UpdateUser godoc
//	@Summary		Update User
//	@Description	Updates an existing User
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer										true	"id"
//	@Param			req	body		types.UpdateUserReq							true	"req"
//	@Success		200	{object}	types.Response{data=types.UpdateUserResp}	"Success"
//	@Failure		400	{object}	types.Response								"Bad request"
//	@Failure		401	{object}	types.Response								"Unauthorized"
//	@Failure		500	{object}	types.Response								"Internal server error"
//	@Security		BearerAuth
//	@Router			/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	var req types.UpdateUserReq
	if err := binding.Bind(c, &req, binding.URI, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.userService.UpdateUser(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("User updated successfully"))
}

// UpdateUserEnabled godoc
//	@Summary		Update User Enabled
//	@Description	Updates an existing User Enabled
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer												true	"id"
//	@Param			req	body		types.UpdateUserEnabledReq							true	"req"
//	@Success		200	{object}	types.Response{data=types.UpdateUserEnabledResp}	"Success"
//	@Failure		400	{object}	types.Response										"Bad request"
//	@Failure		401	{object}	types.Response										"Unauthorized"
//	@Failure		500	{object}	types.Response										"Internal server error"
//	@Security		BearerAuth
//	@Router			/users/{id}/enabled [patch]
func (h *UserHandler) UpdateUserEnabled(c *gin.Context) {
	var req types.UpdateUserEnabledReq
	if err := binding.Bind(c, &req, binding.URI, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.userService.UpdateUserEnabled(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("User enabled status updated successfully"))
}

// DeleteUser godoc
//	@Summary		Delete User
//	@Description	Deletes an existing User
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer										true	"id"
//	@Param			req	body		types.DeleteUserReq							true	"req"
//	@Success		204	{object}	types.Response{data=types.DeleteUserResp}	"Success"
//	@Failure		400	{object}	types.Response								"Bad request"
//	@Failure		401	{object}	types.Response								"Unauthorized"
//	@Failure		500	{object}	types.Response								"Internal server error"
//	@Security		BearerAuth
//	@Router			/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	var req types.DeleteUserReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.userService.DeleteUser(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("User deleted successfully"))
}
