package web

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/handler"
	"github.com/xelarion/go-layout/internal/api/http/web/middleware"
	"github.com/xelarion/go-layout/internal/permission"
)

// Router handles web API routes.
type Router struct {
	logger *zap.Logger

	authMW *jwt.GinJWTMiddleware
	permMW *middleware.PermissionMiddleware

	authHandler       *handler.AuthHandler
	departmentHandler *handler.DepartmentHandler
	permissionHandler *handler.PermissionHandler
	roleHandler       *handler.RoleHandler
	userHandler       *handler.UserHandler
}

// NewRouter creates a new router.
func NewRouter(
	logger *zap.Logger,
	authMW *jwt.GinJWTMiddleware,
	permMW *middleware.PermissionMiddleware,
	authHandler *handler.AuthHandler,
	departmentHandler *handler.DepartmentHandler,
	permissionHandler *handler.PermissionHandler,
	roleHandler *handler.RoleHandler,
	userHandler *handler.UserHandler,
) *Router {
	return &Router{
		logger:            logger.Named("web_router"),
		authMW:            authMW,
		permMW:            permMW,
		authHandler:       authHandler,
		departmentHandler: departmentHandler,
		permissionHandler: permissionHandler,
		roleHandler:       roleHandler,
		userHandler:       userHandler,
	}
}

// Register registers all web API routes to the given router.
func (r *Router) Register(router *gin.Engine) {
	// API routes
	api := router.Group("/api/web/v1")

	// Public routes
	api.POST("/captcha/new", r.authHandler.NewCaptcha)
	api.POST("/captcha/:id/reload", r.authHandler.ReloadCaptcha)
	api.POST("/public_key", r.authHandler.GetRSAPublicKey)
	api.POST("/login", r.authHandler.Login)
	api.GET("/refresh_token", r.authHandler.RefreshToken)

	// Protected routes
	authorized := api.Group("/")
	authorized.Use(r.authMW.MiddlewareFunc())

	authorized.POST("/logout", r.authHandler.Logout)
	authorized.GET("/profile", r.authHandler.GetProfile)
	authorized.PUT("/profile", r.authHandler.UpdateProfile)
	authorized.GET("/users/current", r.authHandler.GetCurrentUserInfo)

	// Department management routes
	authorized.POST("/departments", r.permMW.Check(permission.DepartmentCreate), r.departmentHandler.CreateDepartment)
	authorized.GET("/departments", r.permMW.Check(permission.DepartmentList), r.departmentHandler.ListDepartments)
	authorized.GET("/departments/:id", r.permMW.Check(permission.DepartmentDetail), r.departmentHandler.GetDepartment)
	authorized.GET("/departments/:id/form", r.permMW.Check(permission.DepartmentUpdate), r.departmentHandler.GetDepartmentFormData)
	authorized.PUT("/departments/:id", r.permMW.Check(permission.DepartmentUpdate), r.departmentHandler.UpdateDepartment)
	authorized.PATCH("/departments/:id/enabled", r.permMW.Check(permission.DepartmentUpdateEnabled), r.departmentHandler.UpdateDepartmentEnabled)
	authorized.DELETE("/departments/:id", r.permMW.Check(permission.DepartmentDelete), r.departmentHandler.DeleteDepartment)
	authorized.GET("/departments/options", r.permMW.Check(permission.UserCreate, permission.UserUpdate), r.departmentHandler.GetDepartmentOptions)

	// Role management routes
	authorized.POST("/roles", r.permMW.Check(permission.RoleCreate), r.roleHandler.CreateRole)
	authorized.GET("/roles", r.permMW.Check(permission.RoleList), r.roleHandler.ListRoles)
	authorized.GET("/roles/:id", r.permMW.Check(permission.RoleDetail), r.roleHandler.GetRole)
	authorized.GET("/roles/:id/form", r.permMW.Check(permission.RoleUpdate), r.roleHandler.GetRoleFormData)
	authorized.PUT("/roles/:id", r.permMW.Check(permission.RoleUpdate), r.roleHandler.UpdateRole)
	authorized.PATCH("/roles/:id/enabled", r.permMW.Check(permission.RoleUpdateEnabled), r.roleHandler.UpdateRoleEnabled)
	authorized.DELETE("/roles/:id", r.permMW.Check(permission.RoleDelete), r.roleHandler.DeleteRole)
	authorized.GET("/roles/options", r.permMW.Check(permission.UserCreate, permission.UserUpdate), r.roleHandler.GetRoleOptions)

	// Role permissions routes
	authorized.GET("/roles/:id/permissions", r.permMW.Check(permission.PermissionUpdate), r.permissionHandler.GetRolePermissions)
	authorized.PUT("/roles/:id/permissions", r.permMW.Check(permission.PermissionUpdate), r.permissionHandler.UpdateRolePermissions)
	// Permission tree
	authorized.GET("/permissions/tree", r.permMW.Check(permission.PermissionUpdate), r.permissionHandler.GetPermissionTree)

	// User management routes
	authorized.POST("/users", r.permMW.Check(permission.UserCreate), r.userHandler.CreateUser)
	authorized.GET("/users", r.permMW.Check(permission.UserList), r.userHandler.ListUsers)
	authorized.GET("/users/:id", r.permMW.Check(permission.UserDetail), r.userHandler.GetUser)
	authorized.GET("/users/:id/form", r.permMW.Check(permission.UserUpdate), r.userHandler.GetUserFormData)
	authorized.PUT("/users/:id", r.permMW.Check(permission.UserUpdate), r.userHandler.UpdateUser)
	authorized.PATCH("/users/:id/enabled", r.permMW.Check(permission.UserUpdateEnabled), r.userHandler.UpdateUserEnabled)
	authorized.DELETE("/users/:id", r.permMW.Check(permission.UserDelete), r.userHandler.DeleteUser)
}
