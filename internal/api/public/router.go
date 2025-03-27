// Package public contains Public API handlers and routers for external clients.
package public

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/service"
	"github.com/xelarion/go-layout/pkg/auth"
)

// Router handles all routes for the Public API.
type Router struct {
	Engine      *gin.Engine
	logger      *zap.Logger
	userService *service.UserService
	jwtService  *auth.JWT
}

// NewRouter creates a new Public API router.
func NewRouter(engine *gin.Engine, userService *service.UserService, jwtService *auth.JWT, logger *zap.Logger) *Router {
	return &Router{
		Engine:      engine,
		logger:      logger.Named("public_router"),
		userService: userService,
		jwtService:  jwtService,
	}
}

// SetupRoutes configures all routes for the Public API.
func (r *Router) SetupRoutes() {
	// Initialize handlers
	userHandler := NewUserHandler(r.userService, r.logger)

	// API routes
	api := r.Engine.Group("/public/v1")

	// Auth middleware for public API
	jwtMiddleware := func(c *gin.Context) {
		token := c.GetHeader("X-API-Token")
		if token == "" {
			token = c.Query("api_token")
		}

		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "API token required"})
			return
		}

		claims, err := r.jwtService.VerifyToken(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid or expired API token"})
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}

	// Public routes
	api.POST("/auth", userHandler.Login)

	// Protected routes
	protected := api.Group("/")
	protected.Use(jwtMiddleware)
	{
		protected.GET("/users/:id", userHandler.GetUserInfo)
	}
}
