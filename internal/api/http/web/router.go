package web

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/handler"
	"github.com/xelarion/go-layout/internal/api/http/web/middleware"
	"github.com/xelarion/go-layout/internal/infra/wsocket"
	"github.com/xelarion/go-layout/internal/permission"
)

// Router handles web API routes.
type Router struct {
	logger *zap.Logger

	authMW *jwt.GinJWTMiddleware
	permMW *middleware.PermissionMiddleware

	//	@insert:router_handler_define	请对以下内容进行排序
	announcementHandler *handler.AnnouncementHandler
	authHandler         *handler.AuthHandler
	batchDetailHandler  *handler.BatchDetailHandler
	batchRecordHandler  *handler.BatchRecordHandler
	changeLogHandler    *handler.ChangeLogHandler
	codeHandler         *handler.CodeHandler
	codeRuleHandler     *handler.CodeRuleHandler
	exportHandler       *handler.ExportHandler
	importHandler       *handler.ImportHandler
	notificationHandler *handler.NotificationHandler
	organizationHandler *handler.OrganizationHandler
	permissionHandler   *handler.PermissionHandler
	positionHandler     *handler.PositionHandler
	roleHandler         *handler.RoleHandler
	storageHandler      *handler.StorageHandler
	userHandler         *handler.UserHandler
	wsHandler           *handler.WSHandler
}

// NewRouter creates a new router.
func NewRouter(
	logger *zap.Logger,
	authMW *jwt.GinJWTMiddleware,
	permMW *middleware.PermissionMiddleware,
	_ wsocket.Broadcaster, // 副作用依赖：确保 Broadcaster 被初始化以启动 Redis 订阅

	//	@insert:new_router_handler_define	请对以下内容进行排序
	announcementHandler *handler.AnnouncementHandler,
	authHandler *handler.AuthHandler,
	batchDetailHandler *handler.BatchDetailHandler,
	batchRecordHandler *handler.BatchRecordHandler,
	changeLogHandler *handler.ChangeLogHandler,
	codeHandler *handler.CodeHandler,
	codeRuleHandler *handler.CodeRuleHandler,
	exportHandler *handler.ExportHandler,
	importHandler *handler.ImportHandler,
	notificationHandler *handler.NotificationHandler,
	organizationHandler *handler.OrganizationHandler,
	permissionHandler *handler.PermissionHandler,
	positionHandler *handler.PositionHandler,
	roleHandler *handler.RoleHandler,
	storageHandler *handler.StorageHandler,
	userHandler *handler.UserHandler,
	wsHandler *handler.WSHandler,
) *Router {
	return &Router{
		logger: logger.Named("web_router"),
		authMW: authMW,
		permMW: permMW,

		//	@insert:new_router_handler_return	请对以下内容进行排序
		announcementHandler: announcementHandler,
		authHandler:         authHandler,
		batchDetailHandler:  batchDetailHandler,
		batchRecordHandler:  batchRecordHandler,
		changeLogHandler:    changeLogHandler,
		codeHandler:         codeHandler,
		codeRuleHandler:     codeRuleHandler,
		exportHandler:       exportHandler,
		importHandler:       importHandler,
		notificationHandler: notificationHandler,
		organizationHandler: organizationHandler,
		permissionHandler:   permissionHandler,
		positionHandler:     positionHandler,
		roleHandler:         roleHandler,
		storageHandler:      storageHandler,
		userHandler:         userHandler,
		wsHandler:           wsHandler,
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

	// WebSocket 连接（需要 token 验证）
	api.GET("/ws", r.wsHandler.WSConnect)

	// Protected routes
	authorized := api.Group("/")
	authorized.Use(r.authMW.MiddlewareFunc())

	authorized.POST("/logout", r.authHandler.Logout)
	authorized.GET("/profile", r.authHandler.GetProfile)
	authorized.PUT("/profile", r.authHandler.UpdateProfile)
	authorized.GET("/users/current", r.authHandler.GetCurrentUserInfo)

	// Storage management routes
	authorized.POST("/storage/upload", r.storageHandler.UploadFile)
	authorized.GET("/storage/:key/download_url", r.storageHandler.GetDownloadURL)

	// Import routes
	authorized.GET("/import/template/download", r.importHandler.DownloadImportTemplate)
	authorized.POST("/import/customers", r.permMW.Check(permission.CustomerImport), r.importHandler.ImportCustomers)
	authorized.POST("/import/suppliers", r.permMW.Check(permission.SupplierImport), r.importHandler.ImportSuppliers)
	authorized.POST("/import/equipment_types", r.permMW.Check(permission.EquipmentTypeImport), r.importHandler.ImportEquipmentTypes)
	authorized.POST("/import/equipments", r.permMW.Check(permission.EquipmentImport), r.importHandler.ImportEquipments)

	// Role management routes
	authorized.POST("/roles", r.permMW.Check(permission.RoleCreate), r.roleHandler.CreateRole)
	authorized.GET("/roles", r.permMW.Check(permission.RoleList), r.roleHandler.ListRoles)
	authorized.GET("/roles/:id", r.permMW.Check(permission.RoleDetail, permission.PermissionUpdate), r.roleHandler.GetRole)
	authorized.GET("/roles/:id/form", r.permMW.Check(permission.RoleUpdate), r.roleHandler.GetRoleFormData)
	authorized.PUT("/roles/:id", r.permMW.Check(permission.RoleUpdate), r.roleHandler.UpdateRole)
	authorized.PATCH("/roles/:id/enabled", r.permMW.Check(permission.RoleUpdateEnabled), r.roleHandler.UpdateRoleEnabled)
	authorized.DELETE("/roles/:id", r.permMW.Check(permission.RoleDelete), r.roleHandler.DeleteRole)
	authorized.POST("/roles/batch-delete", r.permMW.Check(permission.RoleDelete), r.roleHandler.BatchDeleteRoles)
	authorized.GET("/roles/options", r.permMW.Check(permission.UserCreate, permission.UserUpdate), r.roleHandler.GetRoleOptions)

	// Role permissions routes
	authorized.GET("/roles/:id/permissions", r.permMW.Check(permission.PermissionUpdate), r.permissionHandler.GetRolePermissions)
	authorized.PUT("/roles/:id/permissions", r.permMW.Check(permission.PermissionUpdate), r.permissionHandler.UpdateRolePermissions)
	// Permission tree
	authorized.GET("/permissions/tree", r.permMW.Check(permission.PermissionUpdate), r.permissionHandler.GetPermissionsTree)

	// Organization management routes
	authorized.POST("/organizations", r.permMW.Check(permission.OrganizationCreate), r.organizationHandler.CreateOrganization)
	authorized.GET("/organizations", r.permMW.Check(permission.OrganizationList), r.organizationHandler.ListOrganizations)
	authorized.GET("/organizations/:id/form", r.permMW.Check(permission.OrganizationUpdate), r.organizationHandler.GetOrganizationFormData)
	authorized.PUT("/organizations/:id", r.permMW.Check(permission.OrganizationUpdate), r.organizationHandler.UpdateOrganization)
	authorized.DELETE("/organizations/:id", r.permMW.Check(permission.OrganizationDelete), r.organizationHandler.DeleteOrganization)
	// Organization tree
	authorized.GET("/organizations/tree", r.permMW.Check(permission.PermissionUpdate), r.organizationHandler.GetOrganizationsTree)

	// Position management routes
	authorized.POST("/positions", r.permMW.Check(permission.PositionCreate), r.positionHandler.CreatePosition)
	authorized.GET("/positions", r.permMW.Check(permission.PositionList), r.positionHandler.ListPositions)
	authorized.GET("/positions/:id/form", r.permMW.Check(permission.PositionUpdate), r.positionHandler.GetPositionFormData)
	authorized.PUT("/positions/:id", r.permMW.Check(permission.PositionUpdate), r.positionHandler.UpdatePosition)
	authorized.DELETE("/positions/:id", r.permMW.Check(permission.PositionDelete), r.positionHandler.DeletePosition)
	authorized.POST("/positions/batch-delete", r.permMW.Check(permission.PositionDelete), r.positionHandler.BatchDeletePositions)
	authorized.GET("/positions/options", r.permMW.Check(permission.UserCreate, permission.UserUpdate), r.positionHandler.GetPositionOptions)

	// Code rules
	authorized.GET("/code_rules", r.permMW.Check(permission.CodeRuleList), r.codeRuleHandler.ListCodeRules)
	authorized.GET("/code_rules/:id/form", r.permMW.Check(permission.CodeRuleUpdate), r.codeRuleHandler.GetCodeRuleFormData)
	authorized.PUT("/code_rules/:id", r.permMW.Check(permission.CodeRuleUpdate), r.codeRuleHandler.UpdateCodeRule)
	authorized.GET("/code_rules/target_form/options", r.permMW.Check(permission.CodeRuleList), r.codeRuleHandler.GetCodeRuleTargetFormOptions)

	// Generate code
	authorized.POST("/codes/generate", r.codeHandler.GenerateCode)
	authorized.POST("/codes/preview", r.codeHandler.PreviewCode)

	// User management routes
	authorized.POST("/users", r.permMW.Check(permission.UserCreate), r.userHandler.CreateUser)
	authorized.GET("/users", r.permMW.Check(permission.UserList), r.userHandler.ListUsers)
	authorized.GET("/users/:id", r.permMW.Check(permission.UserUpdatePassword), r.userHandler.GetUser)
	authorized.GET("/users/:id/form", r.permMW.Check(permission.UserUpdate), r.userHandler.GetUserFormData)
	authorized.PUT("/users/:id", r.permMW.Check(permission.UserUpdate), r.userHandler.UpdateUser)
	authorized.PATCH("/users/:id/enabled", r.permMW.Check(permission.UserUpdateEnabled), r.userHandler.UpdateUserEnabled)
	authorized.PATCH("/users/:id/password", r.permMW.Check(permission.UserUpdatePassword), r.userHandler.UpdateUserPassword)
	authorized.DELETE("/users/:id", r.permMW.Check(permission.UserDelete), r.userHandler.DeleteUser)
	authorized.POST("/users/batch-delete", r.permMW.Check(permission.UserDelete), r.userHandler.BatchDeleteUsers)
	authorized.GET("/users/options", r.permMW.Check(permission.OrganizationCreate, permission.OrganizationUpdate), r.userHandler.GetUserOptions)
	authorized.POST("/users/export", r.permMW.Check(permission.UserExport), r.userHandler.ExportUsers)

	// Export routes
	authorized.GET("/export_tasks", r.exportHandler.ListExportTasks)
	authorized.GET("/export_tasks/:id/download_url", r.exportHandler.GetExportFileDownloadURL)
	authorized.DELETE("/export_tasks/:id", r.exportHandler.DeleteExportTask)

	// Notification management routes
	authorized.GET("/notifications", r.notificationHandler.ListNotifications)
	authorized.GET("/notifications/recent", r.notificationHandler.GetRecentNotifications)
	authorized.GET("/notifications/:id", r.notificationHandler.GetNotification)
	authorized.DELETE("/notifications/:id", r.notificationHandler.DeleteNotification)
	authorized.POST("/notifications/batch-delete", r.notificationHandler.BatchDeleteNotifications)
	authorized.POST("/notifications/batch-read", r.notificationHandler.BatchReadNotifications)
	authorized.POST("/notifications/read-all", r.notificationHandler.ReadAllNotifications)
	// 待办任务统计
	authorized.GET("/todo_stats", r.notificationHandler.GetTodoStats)

	// ChangeLog management routes
	authorized.GET("/change_logs", r.changeLogHandler.ListChangeLogs)

	// Announcement management routes
	authorized.POST("/announcements", r.permMW.Check(permission.AnnouncementCreate), r.announcementHandler.CreateAnnouncement)
	authorized.GET("/announcements", r.permMW.Check(permission.AnnouncementList), r.announcementHandler.ListAnnouncements)
	authorized.GET("/announcements/:id", r.permMW.Check(permission.AnnouncementDetail), r.announcementHandler.GetAnnouncement)
	authorized.GET("/announcements/:id/form", r.permMW.Check(permission.AnnouncementUpdate), r.announcementHandler.GetAnnouncementFormData)
	authorized.PUT("/announcements/:id", r.permMW.Check(permission.AnnouncementUpdate), r.announcementHandler.UpdateAnnouncement)
	authorized.DELETE("/announcements/:id", r.permMW.Check(permission.AnnouncementDelete), r.announcementHandler.DeleteAnnouncement)
	authorized.POST("/announcements/batch-delete", r.permMW.Check(permission.AnnouncementDelete), r.announcementHandler.BatchDeleteAnnouncements)
	authorized.GET("/announcements/publishers/options", r.permMW.Check(permission.AnnouncementList), r.announcementHandler.GetAnnouncementPublishersOptions)

	// Batch record routes
	authorized.GET("/batch_records", r.permMW.Check(permission.BatchRecordList), r.batchRecordHandler.ListBatchRecords)
	authorized.GET("/batch_records/:id", r.permMW.Check(permission.BatchRecordDetail), r.batchRecordHandler.GetBatchRecord)
	authorized.GET("/batch_records/operation_type/options", r.permMW.Check(permission.BatchRecordList), r.batchRecordHandler.GetBatchRecordOperationTypeOptions)
	// Batch detail routes
	authorized.GET("/batch_details", r.permMW.Check(permission.BatchRecordDetail), r.batchDetailHandler.ListBatchDetails)

	//	@insert:router_register	DO NOT DELETE THIS LINE
}
