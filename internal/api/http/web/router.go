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
	Engine            *gin.Engine
	logger            *zap.Logger
	authMW            *jwt.GinJWTMiddleware
	authService       *service.AuthService
	userService       *service.UserService
	departmentService *service.DepartmentService
	roleService       *service.RoleService
}

// NewRouter creates a new router.
func NewRouter(
	engine *gin.Engine,
	authService *service.AuthService,
	userService *service.UserService,
	departmentService *service.DepartmentService,
	roleService *service.RoleService,
	authMiddleware *jwt.GinJWTMiddleware,
	logger *zap.Logger,
) *Router {
	return &Router{
		Engine:            engine,
		logger:            logger.Named("web_router"),
		authMW:            authMiddleware,
		authService:       authService,
		userService:       userService,
		departmentService: departmentService,
		roleService:       roleService,
	}
}

// SetupRoutes configures all routes.
func (r *Router) SetupRoutes() {
	// Initialize handlers
	authHandler := handler.NewAuthHandler(r.authService, r.authMW, r.logger)
	userHandler := handler.NewUserHandler(r.userService, r.logger)
	departmentHandler := handler.NewDepartmentHandler(r.departmentService, r.logger)
	roleHandler := handler.NewRoleHandler(r.roleService, r.logger)

	// API routes
	api := r.Engine.Group("/api/web/v1")

	// Public routes
	api.POST("/captcha/new", authHandler.NewCaptcha)
	api.POST("/captcha/:id/reload", authHandler.ReloadCaptcha)
	api.POST("/public_key", authHandler.GetRSAPublicKey)
	api.POST("/login", authHandler.Login)
	api.GET("/refresh_token", authHandler.RefreshToken)

	// Protected routes - user profile (current user)
	authorized := api.Group("/")
	authorized.Use(r.authMW.MiddlewareFunc())

	authorized.POST("/logout", authHandler.Logout)
	authorized.GET("/profile", authHandler.GetProfile)
	authorized.PUT("/profile", authHandler.UpdateProfile)
	authorized.GET("/users/current", authHandler.GetCurrentUserInfo)

	// User management routes
	authorized.POST("/users", userHandler.CreateUser)
	authorized.GET("/users", userHandler.ListUsers)
	authorized.GET("/users/:id", userHandler.GetUser)
	authorized.GET("/users/:id/form", userHandler.GetUserFormData)
	authorized.PUT("/users/:id", userHandler.UpdateUser)
	authorized.PATCH("/users/:id/enabled", userHandler.UpdateUserEnabled)
	authorized.DELETE("/users/:id", userHandler.DeleteUser)

	// Department management routes
	authorized.POST("/departments", departmentHandler.CreateDepartment)
	authorized.GET("/departments", departmentHandler.ListDepartments)
	authorized.GET("/departments/:id", departmentHandler.GetDepartment)
	authorized.GET("/departments/:id/form", departmentHandler.GetDepartmentFormData)
	authorized.PUT("/departments/:id", departmentHandler.UpdateDepartment)
	authorized.PATCH("/departments/:id/enabled", departmentHandler.UpdateDepartmentEnabled)
	authorized.DELETE("/departments/:id", departmentHandler.DeleteDepartment)
	authorized.GET("/departments/options", departmentHandler.GetDepartmentOptions)

	// Role management routes
	authorized.POST("/roles", roleHandler.CreateRole)
	authorized.GET("/roles", roleHandler.ListRoles)
	authorized.GET("/roles/:id", roleHandler.GetRole)
	authorized.GET("/roles/:id/form", roleHandler.GetRoleFormData)
	authorized.PUT("/roles/:id", roleHandler.UpdateRole)
	authorized.PATCH("/roles/:id/enabled", roleHandler.UpdateRoleEnabled)
	authorized.DELETE("/roles/:id", roleHandler.DeleteRole)
	authorized.GET("/roles/options", roleHandler.GetRoleOptions)
}
