// Package public contains Public API handlers and routers for external clients.
package public

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/service"
)

// Router handles all routes for the Public API.
type Router struct {
	Engine      *gin.Engine
	logger      *zap.Logger
	userService *service.UserService
}

// NewRouter creates a new Public API router.
func NewRouter(engine *gin.Engine, userService *service.UserService, logger *zap.Logger) *Router {
	return &Router{
		Engine:      engine,
		logger:      logger.Named("public_router"),
		userService: userService,
	}
}

// SetupRoutes configures all routes for the Public API.
func (r *Router) SetupRoutes() {
	// Initialize handlers
	userHandler := NewUserHandler(r.userService, r.logger)

	// API routes
	api := r.Engine.Group("/public/v1")

	// API token authentication middleware
	// This is a placeholder that should be implemented by actual users of the template
	apiAuthMiddleware := func(c *gin.Context) {
		// Check for API token in header or query parameter
		token := c.GetHeader("X-API-Token")
		if token == "" {
			token = c.Query("api_token")
		}

		// Placeholder validation - replace with actual validation logic
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"code":    401,
				"message": "API token required",
			})
			return
		}

		// Example: Set user ID after validation
		// In a real implementation, you would validate the token and extract user information
		c.Set("userID", uint(1))
		c.Next()
	}

	// Authentication endpoint - implement your own logic here
	api.POST("/auth", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Implement your authentication logic here",
		})
	})

	// Protected routes - require API token
	protected := api.Group("/")
	protected.Use(apiAuthMiddleware)
	{
		protected.GET("/users/:id", userHandler.GetUserInfo)
	}
}
