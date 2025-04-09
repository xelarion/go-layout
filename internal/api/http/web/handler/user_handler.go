// Package web contains Web API handlers and routers.
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

// UserHandler defines the user HTTP handlers for Web API.
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

// CreateUser handles requests to create a new user.
// @Summary Create a new user
// @Description Create a new user with the provided data
// @Tags users
// @Accept json
// @Produce json
// @Param data body types.CreateUserReq true "User data"
// @Success 201 {object} types.Response{data=types.CreateUserResp} "User created"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 401 {object} types.Response "Unauthorized"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /users [post]
// @Security BearerAuth
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

// ListUsers handles requests to list users with pagination and filtering.
// @Summary List users
// @Description Get a paginated list of users with optional filtering
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param sort_field query string false "Sort field" default(id)
// @Param sort_order query string false "Sort order (asc or desc)" Enums(asc,desc) default(asc)
// @Param username query string false "Filter by username"
// @Param email query string false "Filter by email"
// @Param role query string false "Filter by role"
// @Param enabled query bool false "Filter by enabled status"
// @Success 200 {object} types.Response{data=types.ListUsersResp} "Users list"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 401 {object} types.Response "Unauthorized"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /users [get]
// @Security BearerAuth
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

// GetUser handles requests to get a user by ID.
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

// GetUserFormData provides data needed for user forms (update).
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

// UpdateUser handles requests to update a user.
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

// UpdateUserEnabled handles requests to update a user's enabled status.
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

// DeleteUser handles requests to delete a user.
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

// GetProfile handles requests to get the current user's profile.
func (h *UserHandler) GetProfile(c *gin.Context) {
	var req types.GetProfileReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.userService.GetProfile(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// UpdateProfile handles requests to update the current user's profile.
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req types.UpdateProfileReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		_ = c.Error(errs.WrapValidation(err, err.Error()))
		return
	}

	resp, err := h.userService.UpdateProfile(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("Profile updated successfully"))
}
