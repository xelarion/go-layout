// Package main contains the entry point for the Web API service.
package main

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web"
	"github.com/xelarion/go-layout/internal/api/http/web/middleware"
	"github.com/xelarion/go-layout/internal/api/http/web/service"
	"github.com/xelarion/go-layout/internal/api/http/web/swagger"
	"github.com/xelarion/go-layout/internal/infra/config"
	httpServer "github.com/xelarion/go-layout/internal/infra/server/http"
	"github.com/xelarion/go-layout/internal/repository"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/app"
)

// initApp initializes the Web API application with all needed resources.
// It sets up database connections, creates repositories, services, and middleware.
func initApp(cfgPG *config.PG, cfgRedis *config.Redis, cfgHTTP *config.HTTP, cfgJWT *config.JWT, logger *zap.Logger) (*app.App, func(), error) {
	// Initialize data with connections
	data, dataCleanup, err := repository.NewData(cfgPG, cfgRedis, nil, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize data: %w", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(data)
	departmentRepo := repository.NewDepartmentRepository(data)
	roleRepo := repository.NewRoleRepository(data)

	// Initialize usecases
	userUseCase := usecase.NewUserUseCase(data, userRepo, roleRepo, departmentRepo)
	departmentUseCase := usecase.NewDepartmentUseCase(data, departmentRepo, userRepo)
	roleUseCase := usecase.NewRoleUseCase(data, roleRepo, userRepo)
	permissionUseCase := usecase.NewPermissionUseCase()

	// Initialize services
	departmentService := service.NewDepartmentService(departmentUseCase)
	roleService := service.NewRoleService(roleUseCase)
	userService := service.NewUserService(userUseCase)
	authService := service.NewAuthService(userUseCase, roleUseCase, logger)
	permissionService := service.NewPermissionService(permissionUseCase, roleUseCase)

	// Initialize auth middleware
	authMiddleware, err := middleware.NewAuthMiddleware(cfgJWT, userUseCase, logger)
	if err != nil {
		dataCleanup() // Clean up data resources on error
		return nil, nil, fmt.Errorf("failed to create auth middleware: %w", err)
	}

	permissionMiddleware := middleware.NewPermissionMiddleware(roleUseCase, logger)

	// Initialize HTTP server with options
	hs := httpServer.NewServer(
		cfgHTTP,
		logger,
		httpServer.WithMiddleware(middleware.CORS(cfgHTTP.AllowOrigins)),
		httpServer.WithMiddleware(middleware.Timeout(cfgHTTP.RequestTimeout)),
		httpServer.WithMiddleware(middleware.Recovery(logger)),
		httpServer.WithMiddleware(middleware.Error(logger)),
	)

	// Create web router and register routes
	webRouter := web.NewRouter(
		authService,
		userService,
		departmentService,
		roleService,
		permissionService,
		authMiddleware,
		permissionMiddleware,
		logger,
	)
	hs.RegisterRoutes(webRouter)

	// Swagger router
	swaggerRouter := swagger.NewRouter()
	hs.RegisterRoutes(swaggerRouter)

	// Create application using the newApp function
	appInstance := newApp(logger, hs)

	return appInstance, dataCleanup, nil
}
