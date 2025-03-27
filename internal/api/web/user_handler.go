// Package web contains Web API handlers and routers.
package web

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

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

// Login handles user login requests.
func (h *UserHandler) Login(c *gin.Context) {
	var request struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to service request
	serviceReq := &service.UserRequest{
		Email:    request.Email,
		Password: request.Password,
	}

	response, err := h.userService.LoginUser(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.Warn("Login failed", zap.String("email", request.Email), zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":        response.Token,
		"token_expiry": response.TokenExpiry,
		"user": gin.H{
			"id":    response.User.ID,
			"name":  response.User.Name,
			"email": response.User.Email,
			"role":  response.User.Role,
		},
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

	user, err := h.userService.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get user", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	// Verify user is updating their own profile or is an admin
	userID, _ := c.Get("userID")
	role, _ := c.Get("role")
	if userID.(uint) != uint(id) && role.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own profile"})
		return
	}

	var request struct {
		Name  string `json:"name" binding:"required,min=2,max=100"`
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to service request
	serviceReq := &service.UserRequest{
		ID:    uint(id),
		Name:  request.Name,
		Email: request.Email,
	}

	updatedUser, err := h.userService.UpdateUser(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.Error("Failed to update user", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         updatedUser.ID,
		"name":       updatedUser.Name,
		"email":      updatedUser.Email,
		"role":       updatedUser.Role,
		"updated_at": updatedUser.UpdatedAt,
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

	// Only admin can delete users
	role, _ := c.Get("role")
	if role.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required to delete users"})
		return
	}

	if err := h.userService.DeleteUser(c.Request.Context(), uint(id)); err != nil {
		h.logger.Error("Failed to delete user", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListUsers handles requests to list users with pagination.
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Calculate offset from page and limit
	offset := (page - 1) * limit

	// Create service request
	req := &service.UserRequest{
		Limit:  limit,
		Offset: offset,
	}

	response, err := h.userService.ListUsers(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var userList []gin.H
	for _, user := range response.Users {
		userList = append(userList, gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": userList,
		"meta": gin.H{
			"page":      page,
			"limit":     limit,
			"total":     response.Count,
			"last_page": (response.Count + limit - 1) / limit,
		},
	})
}

// GetProfile handles requests to get the current user's profile.
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	user, err := h.userService.GetUser(c.Request.Context(), userID.(uint))
	if err != nil {
		h.logger.Error("Failed to get user profile", zap.Uint("id", userID.(uint)), zap.Error(err))
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
