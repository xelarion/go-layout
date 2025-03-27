// Package public contains Public API handlers and routers for external clients.
package public

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/service"
)

// UserHandler defines the user HTTP handlers for Public API.
type UserHandler struct {
	userService *service.UserService
	logger      *zap.Logger
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(userService *service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger.Named("public_user_handler"),
	}
}

// Login handles user login requests for public API.
func (h *UserHandler) Login(c *gin.Context) {
	var request struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
		ApiKey   string `json:"api_key" binding:"required"`
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
		h.logger.Warn("Public API login failed", zap.String("email", request.Email), zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": response.Token,
		"expires_at":   response.TokenExpiry,
		"user_info": gin.H{
			"user_id": response.User.ID,
			"name":    response.User.Name,
		},
	})
}

// GetUserInfo handles requests to get a user's public information.
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get user info", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Return limited user information for public API
	c.JSON(http.StatusOK, gin.H{
		"user_id":   user.ID,
		"name":      user.Name,
		"joined_at": user.CreatedAt.Format("2006-01-02"),
	})
}
