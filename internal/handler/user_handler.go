// Package handler contains HTTP request handlers.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/service"
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
		logger:      logger.Named("user_handler"),
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

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ListUsers handles requests to list users with pagination.
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Get pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Create service request with pagination
	serviceReq := &service.UserRequest{
		Limit:  limit,
		Offset: offset,
	}

	response, err := h.userService.ListUsers(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Int("limit", limit), zap.Int("offset", offset), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Transform to response format
	var userResponseList []gin.H
	for _, user := range response.Users {
		userResponseList = append(userResponseList, gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  userResponseList,
		"limit":  response.Limit,
		"offset": response.Offset,
		"count":  response.Count,
	})
}

// GetProfile handles requests to get the current user's profile.
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

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

// RegisterRoutes registers the user routes on the given router group.
func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc, adminMiddleware gin.HandlerFunc) {
	// Public routes
	router.POST("/register", h.Register)
	router.POST("/login", h.Login)

	// Protected routes
	protected := router.Group("/")
	protected.Use(authMiddleware)
	{
		protected.GET("/profile", h.GetProfile)
		protected.GET("/users/:id", h.GetUser)
		protected.PUT("/users/:id", h.UpdateUser)
	}

	// Admin routes
	admin := protected.Group("/")
	admin.Use(adminMiddleware)
	{
		admin.GET("/users", h.ListUsers)
		admin.DELETE("/users/:id", h.DeleteUser)
	}
}
