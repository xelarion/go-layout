// Package web contains Web API handlers and routers.
package web

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/service"
	"github.com/xelarion/go-layout/pkg/auth"
)

// Router handles all routes for the Web API.
type Router struct {
	Engine      *gin.Engine
	logger      *zap.Logger
	userService *service.UserService
	jwtService  *auth.JWT
}

// NewRouter creates a new Web API router.
func NewRouter(engine *gin.Engine, userService *service.UserService, jwtService *auth.JWT, logger *zap.Logger) *Router {
	return &Router{
		Engine:      engine,
		logger:      logger.Named("web_router"),
		userService: userService,
		jwtService:  jwtService,
	}
}

// SetupRoutes configures all routes for the Web API.
func (r *Router) SetupRoutes() {
	// Initialize handlers
	userHandler := NewUserHandler(r.userService, r.logger)

	// API routes
	api := r.Engine.Group("/api/v1")

	// Auth middleware
	jwtMiddleware := func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			token = c.Query("token")
		}

		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Authentication required"})
			return
		}

		claims, err := r.jwtService.VerifyToken(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid or expired token"})
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}

	// Admin middleware
	adminMiddleware := func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(401, gin.H{"error": "Authentication required"})
			return
		}

		if role.(string) != "admin" {
			c.AbortWithStatusJSON(403, gin.H{"error": "Admin access required"})
			return
		}

		c.Next()
	}

	// Public routes
	api.POST("/register", userHandler.Register)
	api.POST("/login", userHandler.Login)

	// Protected routes
	protected := api.Group("/")
	protected.Use(jwtMiddleware)
	{
		protected.GET("/profile", userHandler.GetProfile)
		protected.GET("/users/:id", userHandler.GetUser)
		protected.PUT("/users/:id", userHandler.UpdateUser)
	}

	// Admin routes
	admin := protected.Group("/")
	admin.Use(adminMiddleware)
	{
		admin.GET("/users", userHandler.ListUsers)
		admin.DELETE("/users/:id", userHandler.DeleteUser)
	}
}
