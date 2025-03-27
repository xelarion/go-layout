// Package web contains Web API handlers and routers.
package web

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/web/middleware"
	"github.com/xelarion/go-layout/internal/service"
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

// Register handles user registration requests.
func (h *UserHandler) Register(c *gin.Context) {
	var request struct {
		Name     string `json:"name" binding:"required,min=2,max=100"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to service request
	serviceReq := &service.UserRequest{
		Name:     request.Name,
		Email:    request.Email,
		Password: request.Password,
	}

	user, err := h.userService.RegisterUser(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.Error("Failed to register user", zap.String("email", request.Email), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
	})
}

// GetUser handles requests to get a user by ID.
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get current authenticated user for permission check
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Only allow users to view their own profile or admin to view any profile
	if currentUser.ID != uint(id) && !currentUser.IsAdmin() {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get user", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"role":       user.Role,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	})
}

// UpdateUser handles requests to update a user.
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get current authenticated user for permission check
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Only allow users to update their own profile or admin to update any profile
	if currentUser.ID != uint(id) && !currentUser.IsAdmin() {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var request struct {
		Name     string `json:"name" binding:"omitempty,min=2,max=100"`
		Email    string `json:"email" binding:"omitempty,email"`
		Password string `json:"password" binding:"omitempty,min=6"`
		Role     string `json:"role" binding:"omitempty,oneof=user admin"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Only admin can update role
	if request.Role != "" && !currentUser.IsAdmin() {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can update roles"})
		return
	}

	// Convert to service request
	serviceReq := &service.UserRequest{
		ID:       uint(id),
		Name:     request.Name,
		Email:    request.Email,
		Password: request.Password,
		Role:     request.Role,
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.Error("Failed to update user", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"role":       user.Role,
		"updated_at": user.UpdatedAt,
	})
}

// DeleteUser handles requests to delete a user.
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Admin check is handled by middleware in the router.

	err = h.userService.DeleteUser(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to delete user", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// ListUsers handles requests to list users with pagination.
func (h *UserHandler) ListUsers(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset value"})
		return
	}

	// Admin check is handled by middleware in the router.

	serviceReq := &service.UserRequest{
		Limit:  limit,
		Offset: offset,
	}

	listResponse, err := h.userService.ListUsers(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  listResponse.Users,
		"limit":  listResponse.Limit,
		"offset": listResponse.Offset,
		"count":  listResponse.Count,
	})
}

// GetProfile handles requests to get the current user's profile.
func (h *UserHandler) GetProfile(c *gin.Context) {
	// Get current authenticated user from context
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), currentUser.ID)
	if err != nil {
		h.logger.Error("Failed to get user profile", zap.Uint("id", currentUser.ID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"role":       user.Role,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	})
}
