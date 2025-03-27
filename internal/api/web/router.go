// Package web contains Web API handlers and routers.
package web

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/web/middleware"
	"github.com/xelarion/go-layout/internal/service"
)

// Router handles all routes for the Web API.
type Router struct {
	Engine      *gin.Engine
	logger      *zap.Logger
	userService *service.UserService
	authMW      *jwt.GinJWTMiddleware
}

// NewRouter creates a new Web API router.
func NewRouter(engine *gin.Engine, userService *service.UserService, authMiddleware *jwt.GinJWTMiddleware, logger *zap.Logger) *Router {
	return &Router{
		Engine:      engine,
		logger:      logger.Named("web_router"),
		userService: userService,
		authMW:      authMiddleware,
	}
}

// SetupRoutes configures all routes for the Web API.
func (r *Router) SetupRoutes() {
	// Initialize handlers
	userHandler := NewUserHandler(r.userService, r.logger)

	// API routes
	api := r.Engine.Group("/api/v1")

	// Public routes
	api.POST("/auth", r.authMW.LoginHandler)
	api.GET("/refresh_token", r.authMW.RefreshHandler)
	api.POST("/register", userHandler.Register)

	// Protected routes
	protected := api.Group("/")
	protected.Use(r.authMW.MiddlewareFunc())
	{
		protected.GET("/profile", userHandler.GetProfile)
		protected.GET("/users/:id", userHandler.GetUser)
		protected.PUT("/users/:id", userHandler.UpdateUser)
		
		// Admin routes - require admin role
		admin := protected.Group("/admin")
		admin.Use(middleware.AdminAuthorizatorMiddleware())
		{
			admin.GET("/users", userHandler.ListUsers)
			admin.DELETE("/users/:id", userHandler.DeleteUser)
		}
	}
}
