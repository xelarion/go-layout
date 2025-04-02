// Package web contains Web API handlers and routers.
package web

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/web/handler"
	"github.com/xelarion/go-layout/internal/api/web/middleware"
	"github.com/xelarion/go-layout/internal/service"
)

// Router handles all routes for the Web API.
type Router struct {
	Engine      *gin.Engine
	logger      *zap.Logger
	userService *service.UserService
	authService *service.AuthService
	authMW      *jwt.GinJWTMiddleware
}

// NewRouter creates a new Web API router.
func NewRouter(engine *gin.Engine, userService *service.UserService, authService *service.AuthService, authMiddleware *jwt.GinJWTMiddleware, logger *zap.Logger) *Router {
	return &Router{
		Engine:      engine,
		logger:      logger.Named("web_router"),
		userService: userService,
		authService: authService,
		authMW:      authMiddleware,
	}
}

// SetupRoutes configures all routes for the Web API.
func (r *Router) SetupRoutes() {
	// Initialize handlers
	userHandler := handler.NewUserHandler(r.userService, r.logger)
	authHandler := handler.NewAuthHandler(r.authService, r.logger)

	// API routes
	api := r.Engine.Group("/api/v1")

	// Public routes
	api.POST("/login", r.authMW.LoginHandler)
	api.GET("/refresh_token", r.authMW.RefreshHandler)
	api.GET("/captcha", authHandler.GetCaptcha)

	// Protected routes - user profile (current user)
	authorized := api.Group("/")
	authorized.Use(r.authMW.MiddlewareFunc())

	// User profile routes
	authorized.GET("/profile", userHandler.GetProfile)
	authorized.PUT("/profile", userHandler.UpdateProfile)

	adminOnly := authorized.Group("/")
	adminOnly.Use(middleware.AdminOnly())

	// User management routes (admin only)
	api.POST("/users", userHandler.CreateUser)
	api.GET("/users", userHandler.ListUsers)
	api.GET("/users/:id", userHandler.GetUser)
	api.GET("/users/:id/form", userHandler.GetUserFormData)
	api.PUT("/users/:id", userHandler.UpdateUser)
	api.PATCH("/users/:id/enabled", userHandler.UpdateUserEnabled)
	api.DELETE("/users/:id", userHandler.DeleteUser)
}
