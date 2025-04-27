//go:build wireinject
// +build wireinject

// Package main contains the wire implementation for dependency injection.
package main

import (
	"github.com/google/wire"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web"
	"github.com/xelarion/go-layout/internal/api/http/web/middleware"
	"github.com/xelarion/go-layout/internal/api/http/web/service"
	"github.com/xelarion/go-layout/internal/api/http/web/swagger"
	"github.com/xelarion/go-layout/internal/api/http/web/handler"
	"github.com/xelarion/go-layout/internal/infra/config"
	httpServer "github.com/xelarion/go-layout/internal/infra/server/http"
	"github.com/xelarion/go-layout/internal/repository"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/app"
)

// initApp initializes the Web API application with all needed resources.
// It sets up database connections, creates repositories, services, and middleware.
func initApp(cfgPG *config.PG, cfgRedis *config.Redis, cfgRabbitMQ *config.RabbitMQ, cfgHTTP *config.HTTP, cfgJWT *config.JWT, logger *zap.Logger) (*app.App, func(), error) {
	wire.Build(
		repository.ProviderSet,
		usecase.ProviderSet,
		service.ProviderSet,
		handler.ProviderSet,
		middleware.NewAuthMiddleware,
		middleware.NewPermissionMiddleware,
		web.NewRouter,
		swagger.NewRouter,
		provideHTTPServer,
		newApp,
	)
	return nil, nil, nil
}

// provideHTTPServer provides the HTTP server instance.
func provideHTTPServer(cfgHTTP *config.HTTP, logger *zap.Logger, webRouter *web.Router, swaggerRouter *swagger.Router) *httpServer.Server {
	hs := httpServer.NewServer(
		cfgHTTP,
		logger,
		httpServer.WithMiddleware(middleware.CORS(cfgHTTP.AllowOrigins)),
		httpServer.WithMiddleware(middleware.Timeout(cfgHTTP.RequestTimeout)),
		httpServer.WithMiddleware(middleware.Recovery(logger)),
		httpServer.WithMiddleware(middleware.Error(logger)),
	)

	hs.RegisterRoutes(webRouter)
	hs.RegisterRoutes(swaggerRouter)

	return hs
}
