// Package web contains Web API handlers and routers.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/web/middleware"
	"github.com/xelarion/go-layout/internal/api/web/types"
	"github.com/xelarion/go-layout/internal/service"
	"github.com/xelarion/go-layout/pkg/binding"
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
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req types.CreateUserReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, types.Error(types.CodeValidation, err.Error()))
		return
	}

	resp, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create user", zap.String("email", req.Email), zap.Error(err))
		c.JSON(http.StatusInternalServerError, types.Error(types.CodeInternalError, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, types.Success(resp).WithMessage("User created successfully"))
}

// ListUsers handles requests to list users with pagination and filtering.
func (h *UserHandler) ListUsers(c *gin.Context) {
	var req types.ListUsersReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		c.JSON(http.StatusBadRequest, types.Error(types.CodeValidation, err.Error()))
		return
	}

	resp, err := h.userService.ListUsers(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, types.Error(types.CodeInternalError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// GetUser handles requests to get a user by ID.
func (h *UserHandler) GetUser(c *gin.Context) {
	var req types.GetUserReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		c.JSON(http.StatusBadRequest, types.Error(types.CodeValidation, err.Error()))
		return
	}

	resp, err := h.userService.GetUser(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to get user", zap.String("id", c.Param("id")), zap.Error(err))
		c.JSON(http.StatusInternalServerError, types.Error(types.CodeInternalError, "Internal server error"))
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
		c.JSON(http.StatusBadRequest, types.Error(types.CodeValidation, err.Error()))
		return
	}

	resp, err := h.userService.GetUserFormData(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to get user", zap.String("id", c.Param("id")), zap.Error(err))
		c.JSON(http.StatusInternalServerError, types.Error(types.CodeInternalError, "Internal server error"))
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// UpdateUser handles requests to update a user.
func (h *UserHandler) UpdateUser(c *gin.Context) {
	var req types.UpdateUserReq
	if err := binding.Bind(c, &req, binding.URI, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, types.Error(types.CodeValidation, err.Error()))
		return
	}

	resp, err := h.userService.UpdateUser(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update user", zap.String("id", c.Param("id")), zap.Error(err))
		c.JSON(http.StatusInternalServerError, types.Error(types.CodeInternalError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("User updated successfully"))
}

// UpdateUserEnabled handles requests to update a user's enabled status.
func (h *UserHandler) UpdateUserEnabled(c *gin.Context) {
	var req types.UpdateUserEnabledReq
	if err := binding.Bind(c, &req, binding.URI, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, types.Error(types.CodeValidation, err.Error()))
		return
	}

	resp, err := h.userService.UpdateUserEnabled(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update user enabled status", zap.String("id", c.Param("id")), zap.Error(err))
		c.JSON(http.StatusInternalServerError, types.Error(types.CodeInternalError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("User enabled status updated successfully"))
}

// DeleteUser handles requests to delete a user.
func (h *UserHandler) DeleteUser(c *gin.Context) {
	var req types.DeleteUserReq
	if err := binding.Bind(c, &req, binding.URI, binding.Query); err != nil {
		c.JSON(http.StatusBadRequest, types.Error(types.CodeValidation, err.Error()))
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
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, types.Error(types.CodeUnauthorized, "Unauthorized"))
		return
	}

	var req types.GetProfileReq
	if err := binding.Bind(c, &req, binding.Query); err != nil {
		c.JSON(http.StatusBadRequest, types.Error(types.CodeValidation, err.Error()))
		return
	}

	resp, err := h.userService.GetProfile(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to get user profile", zap.Uint("id", currentUser.ID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, types.Error(types.CodeInternalError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, types.Success(resp))
}

// UpdateProfile handles requests to update the current user's profile.
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, types.Error(types.CodeUnauthorized, "Unauthorized"))
		return
	}

	var req types.UpdateProfileReq
	if err := binding.Bind(c, &req, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, types.Error(types.CodeValidation, err.Error()))
		return
	}

	resp, err := h.userService.UpdateProfile(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update profile", zap.Uint("id", currentUser.ID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, types.Error(types.CodeInternalError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, types.Success(resp).WithMessage("Profile updated successfully"))
}
