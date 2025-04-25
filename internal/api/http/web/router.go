package web

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/handler"
	"github.com/xelarion/go-layout/internal/api/http/web/middleware"
	"github.com/xelarion/go-layout/internal/api/http/web/service"
	"github.com/xelarion/go-layout/internal/permission"
)

// Router handles web API routes.
type Router struct {
	logger            *zap.Logger
	authMW            *jwt.GinJWTMiddleware
	permMW            *middleware.PermissionMiddleware
	authService       *service.AuthService
	userService       *service.UserService
	departmentService *service.DepartmentService
	roleService       *service.RoleService
	permissionService *service.PermissionService
}

// NewRouter creates a new router.
func NewRouter(
	authService *service.AuthService,
	userService *service.UserService,
	departmentService *service.DepartmentService,
	roleService *service.RoleService,
	permissionService *service.PermissionService,
	authMiddleware *jwt.GinJWTMiddleware,
	permissionMiddleware *middleware.PermissionMiddleware,
	logger *zap.Logger,
) *Router {
	return &Router{
		logger:            logger.Named("web_router"),
		authMW:            authMiddleware,
		permMW:            permissionMiddleware,
		authService:       authService,
		userService:       userService,
		departmentService: departmentService,
		roleService:       roleService,
		permissionService: permissionService,
	}
}

// Register registers all web API routes to the given router.
func (r *Router) Register(router *gin.Engine) {
	// Initialize handlers
	authHandler := handler.NewAuthHandler(r.authService, r.authMW, r.logger)
	userHandler := handler.NewUserHandler(r.userService, r.logger)
	departmentHandler := handler.NewDepartmentHandler(r.departmentService, r.logger)
	roleHandler := handler.NewRoleHandler(r.roleService, r.logger)
	permissionHandler := handler.NewPermissionHandler(r.permissionService, r.logger)

	// API routes
	api := router.Group("/api/web/v1")

	// Public routes
	api.POST("/captcha/new", authHandler.NewCaptcha)
	api.POST("/captcha/:id/reload", authHandler.ReloadCaptcha)
	api.POST("/public_key", authHandler.GetRSAPublicKey)
	api.POST("/login", authHandler.Login)
	api.GET("/refresh_token", authHandler.RefreshToken)

	// Protected routes
	authorized := api.Group("/")
	authorized.Use(r.authMW.MiddlewareFunc())

	authorized.POST("/logout", authHandler.Logout)
	authorized.GET("/profile", authHandler.GetProfile)
	authorized.PUT("/profile", authHandler.UpdateProfile)
	authorized.GET("/users/current", authHandler.GetCurrentUserInfo)

	// Department management routes
	authorized.POST("/departments", r.permMW.Check(permission.DepartmentCreate), departmentHandler.CreateDepartment)
	authorized.GET("/departments", r.permMW.Check(permission.DepartmentList), departmentHandler.ListDepartments)
	authorized.GET("/departments/:id", r.permMW.Check(permission.DepartmentDetail), departmentHandler.GetDepartment)
	authorized.GET("/departments/:id/form", r.permMW.Check(permission.DepartmentUpdate), departmentHandler.GetDepartmentFormData)
	authorized.PUT("/departments/:id", r.permMW.Check(permission.DepartmentUpdate), departmentHandler.UpdateDepartment)
	authorized.PATCH("/departments/:id/enabled", r.permMW.Check(permission.DepartmentUpdateEnabled), departmentHandler.UpdateDepartmentEnabled)
	authorized.DELETE("/departments/:id", r.permMW.Check(permission.DepartmentDelete), departmentHandler.DeleteDepartment)
	authorized.GET("/departments/options", r.permMW.Check(permission.UserCreate, permission.UserUpdate), departmentHandler.GetDepartmentOptions)

	// Role management routes
	authorized.POST("/roles", r.permMW.Check(permission.RoleCreate), roleHandler.CreateRole)
	authorized.GET("/roles", r.permMW.Check(permission.RoleList), roleHandler.ListRoles)
	authorized.GET("/roles/:id", r.permMW.Check(permission.RoleDetail), roleHandler.GetRole)
	authorized.GET("/roles/:id/form", r.permMW.Check(permission.RoleUpdate), roleHandler.GetRoleFormData)
	authorized.PUT("/roles/:id", r.permMW.Check(permission.RoleUpdate), roleHandler.UpdateRole)
	authorized.PATCH("/roles/:id/enabled", r.permMW.Check(permission.RoleUpdateEnabled), roleHandler.UpdateRoleEnabled)
	authorized.DELETE("/roles/:id", r.permMW.Check(permission.RoleDelete), roleHandler.DeleteRole)
	authorized.GET("/roles/options", r.permMW.Check(permission.UserCreate, permission.UserUpdate), roleHandler.GetRoleOptions)
	// Role permissions routes
	authorized.GET("/roles/:id/permissions", r.permMW.Check(permission.PermissionUpdate), permissionHandler.GetRolePermissions)
	authorized.PUT("/roles/:id/permissions", r.permMW.Check(permission.PermissionUpdate), permissionHandler.UpdateRolePermissions)
	// Permission tree
	authorized.GET("/permissions/tree", r.permMW.Check(permission.PermissionUpdate), permissionHandler.GetPermissionTree)

	// User management routes
	authorized.POST("/users", r.permMW.Check(permission.UserCreate), userHandler.CreateUser)
	authorized.GET("/users", r.permMW.Check(permission.UserList), userHandler.ListUsers)
	authorized.GET("/users/:id", r.permMW.Check(permission.UserDetail), userHandler.GetUser)
	authorized.GET("/users/:id/form", r.permMW.Check(permission.UserUpdate), userHandler.GetUserFormData)
	authorized.PUT("/users/:id", r.permMW.Check(permission.UserUpdate), userHandler.UpdateUser)
	authorized.PATCH("/users/:id/enabled", r.permMW.Check(permission.UserUpdateEnabled), userHandler.UpdateUserEnabled)
	authorized.DELETE("/users/:id", r.permMW.Check(permission.UserDelete), userHandler.DeleteUser)
}
