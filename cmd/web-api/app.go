package main

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web"
	"github.com/xelarion/go-layout/internal/api/http/web/middleware"
	"github.com/xelarion/go-layout/internal/api/http/web/service"
	"github.com/xelarion/go-layout/internal/api/http/web/swagger"
	"github.com/xelarion/go-layout/internal/cache"
	"github.com/xelarion/go-layout/internal/config"
	"github.com/xelarion/go-layout/internal/database"
	"github.com/xelarion/go-layout/internal/repository"
	"github.com/xelarion/go-layout/internal/server"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/app"
)

// initApp initializes the application with all needed components.
// It sets up the database connection, repositories, usecases, services,
// and HTTP server with all routes.
func initApp(cfgPG *config.PG, cfgRedis *config.Redis, cfgHTTP *config.HTTP, cfgJWT *config.JWT, logger *zap.Logger) (*app.App, error) {
	logger.Info("Initializing Web API application")

	// Initialize database connection
	db, err := database.NewPostgres(cfgPG, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	logger.Info("Connected to database successfully")

	// Initialize redis connection
	redis, err := cache.NewRedis(cfgRedis, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	logger.Info("Connected to redis successfully")

	// Initialize data
	data := repository.NewData(db.DB, redis.Client)
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
		return nil, fmt.Errorf("failed to create auth middleware: %w", err)
	}

	permissionMiddleware := middleware.NewPermissionMiddleware(roleUseCase, logger)

	// Initialize HTTP server
	httpServer := server.NewHTTPServer(cfgHTTP, logger)

	// Setup global middlewares
	httpServer.Router().Use(middleware.CORS(cfgHTTP.AllowOrigins))
	httpServer.Router().Use(middleware.Timeout(cfgHTTP.RequestTimeout))
	httpServer.Router().Use(middleware.Recovery(logger))
	httpServer.Router().Use(middleware.Error(logger))

	// Register API routes
	webRouter := web.NewRouter(
		httpServer.Router(),
		authService,
		userService,
		departmentService,
		roleService,
		permissionService,
		authMiddleware,
		permissionMiddleware,
		logger,
	)
	webRouter.SetupRoutes()

	// Register Swagger documentation routes (when swag is installed)
	swagger.RegisterRoutes(httpServer.Router())

	// Create the application with start and stop functions
	logger.Info("Creating Web API application")
	apiApp := app.NewApp(
		"web-api",
		"Web API Service",
		"0.1.0",
		logger,
		app.WithStartFunc(func(ctx context.Context) error {
			logger.Info("Starting Web API server",
				zap.String("host", cfgHTTP.Host),
				zap.Int("port", cfgHTTP.Port),
				zap.String("mode", cfgHTTP.Mode))

			// Start the HTTP server
			if err := httpServer.Start(); err != nil {
				return fmt.Errorf("failed to start HTTP server: %w", err)
			}
			return nil
		}),
		app.WithStopFunc(func(ctx context.Context) error {
			logger.Info("Stopping Web API server")

			// Create a timeout context for graceful shutdown
			shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			// Shutdown the HTTP server
			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				logger.Error("Failed to shutdown HTTP server gracefully", zap.Error(err))
				return err
			} else {
				logger.Info("HTTP server shutdown successfully")
			}

			// Close database connection
			if err := db.Close(); err != nil {
				logger.Error("Error closing database connection", zap.Error(err))
			} else {
				logger.Info("Database connection closed successfully")
			}

			// Close redis connection
			if err := redis.Close(); err != nil {
				logger.Error("Error closing redis connection", zap.Error(err))
			} else {
				logger.Info("Redis connection closed successfully")
			}

			logger.Info("Web API server stopped successfully")
			return nil
		}),
	)

	logger.Info("Web API application initialized successfully")
	return apiApp, nil
}
