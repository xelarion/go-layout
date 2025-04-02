// Package main contains the entry point for the API service.
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/public"
	"github.com/xelarion/go-layout/internal/api/web"
	"github.com/xelarion/go-layout/internal/api/web/middleware"
	"github.com/xelarion/go-layout/internal/repository"
	"github.com/xelarion/go-layout/internal/service"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/app"
	"github.com/xelarion/go-layout/pkg/config"
	"github.com/xelarion/go-layout/pkg/database"
	"github.com/xelarion/go-layout/pkg/server"
)

// initApp initializes the API application with all needed components.
// It sets up the database connection, repositories, usecases, services,
// and HTTP server with all routes.
func initApp(cfg *config.Config, logger *zap.Logger) (*app.App, error) {
	logger.Info("Initializing API application")

	// Initialize database connection
	db, err := database.NewPostgres(&cfg.PG, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	logger.Info("Connected to database successfully")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.DB, logger)
	// Initialize usecases
	userUseCase := usecase.NewUserUseCase(userRepo, logger)
	// Initialize services (without JWT dependency)
	userService := service.NewUserService(userUseCase)

	// Initialize auth middleware
	authMiddleware, err := middleware.NewAuthMiddleware(&cfg.JWT, userUseCase, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth middleware: %w", err)
	}

	// Initialize HTTP server
	httpServer := server.NewHTTPServer(&server.HTTPConfig{
		Host:         cfg.HTTP.Host,
		Port:         cfg.HTTP.Port,
		Mode:         cfg.HTTP.Mode,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}, logger)

	// Setup global middlewares
	httpServer.Router().Use(middleware.Logger(logger))
	// Add recovery middleware to recover from panics
	httpServer.Router().Use(gin.Recovery())
	// Add request timeout middleware
	httpServer.Router().Use(middleware.Timeout(30 * time.Second))

	// Initialize auth service
	authService := service.NewAuthService(userUseCase, logger)

	// Register Web API routes
	webRouter := web.NewRouter(httpServer.Router(), userService, authService, authMiddleware, logger)
	webRouter.SetupRoutes()

	// Register Public API routes
	publicRouter := public.NewRouter(httpServer.Router(), userService, logger)
	publicRouter.SetupRoutes()

	// Common health check endpoint
	logger.Debug("Setting up health check endpoint")
	httpServer.Router().GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "api",
			"version": "0.1.0",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// Create the application with start and stop functions
	logger.Info("Creating API application")
	apiApp := app.NewApp(
		"api",
		"API Service",
		"0.1.0",
		logger,
		app.WithStartFunc(func(ctx context.Context) error {
			logger.Info("Starting API server",
				zap.String("host", cfg.HTTP.Host),
				zap.Int("port", cfg.HTTP.Port),
				zap.String("mode", cfg.HTTP.Mode))

			// Start the HTTP server
			if err := httpServer.Start(); err != nil {
				return fmt.Errorf("failed to start HTTP server: %w", err)
			}
			return nil
		}),
		app.WithStopFunc(func(ctx context.Context) error {
			logger.Info("Stopping API server")

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

			logger.Info("API server stopped successfully")
			return nil
		}),
	)

	logger.Info("API application initialized successfully")
	return apiApp, nil
}
