package web

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/handler"
	"github.com/xelarion/go-layout/internal/api/http/web/service"
)

// Router handles all routes.
type Router struct {
	Engine      *gin.Engine
	logger      *zap.Logger
	userService *service.UserService
	authService *service.AuthService
	authMW      *jwt.GinJWTMiddleware
}

// NewRouter creates a new router.
func NewRouter(engine *gin.Engine, userService *service.UserService, authService *service.AuthService, authMiddleware *jwt.GinJWTMiddleware, logger *zap.Logger) *Router {
	return &Router{
		Engine:      engine,
		logger:      logger.Named("web_router"),
		userService: userService,
		authService: authService,
		authMW:      authMiddleware,
	}
}

// SetupRoutes configures all routes.
func (r *Router) SetupRoutes() {
	// Initialize handlers
	userHandler := handler.NewUserHandler(r.userService, r.logger)
	authHandler := handler.NewAuthHandler(r.authService, r.logger)

	// API routes
	api := r.Engine.Group("/api/web")

	// Public routes
	api.POST("/captcha/new", authHandler.NewCaptcha)
	api.POST("/captcha/:id/reload", authHandler.ReloadCaptcha)
	api.POST("/public_key", authHandler.GetRSAPublicKey)
	api.POST("/login", r.authMW.LoginHandler)
	api.GET("/refresh_token", r.authMW.RefreshHandler)

	// Protected routes - user profile (current user)
	authorized := api.Group("/")
	authorized.Use(r.authMW.MiddlewareFunc())

	authorized.GET("/users/current", authHandler.GetCurrentUserInfo)
	authorized.POST("/logout", r.authMW.LogoutHandler)

	// User profile routes
	authorized.GET("/profile", userHandler.GetProfile)
	authorized.PUT("/profile", userHandler.UpdateProfile)

	// User management routes
	authorized.POST("/users", userHandler.CreateUser)
	authorized.GET("/users", userHandler.ListUsers)
	authorized.GET("/users/:id", userHandler.GetUser)
	authorized.GET("/users/:id/form", userHandler.GetUserFormData)
	authorized.PUT("/users/:id", userHandler.UpdateUser)
	authorized.PATCH("/users/:id/enabled", userHandler.UpdateUserEnabled)
	authorized.DELETE("/users/:id", userHandler.DeleteUser)
}
