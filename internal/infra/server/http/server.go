// Package http provides HTTP server implementation for the application.
package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/xelarion/go-layout/internal/infra/config"
)

type zapWriter struct {
	logger *zap.Logger
	level  zapcore.Level
}

func (w *zapWriter) Write(p []byte) (n int, err error) {
	switch w.level {
	case zapcore.ErrorLevel:
		w.logger.Error(string(p))
	default:
		w.logger.Info(string(p))
	}
	return len(p), nil
}

// Server represents an HTTP server.
type Server struct {
	server *http.Server
	router *gin.Engine
	logger *zap.Logger
}

// RouterRegistrar is an interface for router registration.
type RouterRegistrar interface {
	Register(router *gin.Engine)
}

// ServerOption is a function that configures a Server.
type ServerOption func(*Server)

// WithMiddleware adds middleware to the router.
func WithMiddleware(middleware gin.HandlerFunc) ServerOption {
	return func(s *Server) {
		s.router.Use(middleware)
	}
}

// NewServer creates a new HTTP server instance.
func NewServer(config *config.HTTP, logger *zap.Logger, opts ...ServerOption) *Server {
	// Set Gin mode
	gin.SetMode(config.Mode)

	// disable console color in production mode
	if config.Mode == gin.ReleaseMode {
		gin.DisableConsoleColor()
	}

	// Set Gin logger
	httpLogger := logger.WithOptions(
		zap.WithCaller(false),
		zap.AddStacktrace(zap.FatalLevel),
	)
	gin.DefaultWriter = &zapWriter{logger: httpLogger, level: zapcore.InfoLevel}
	gin.DefaultErrorWriter = &zapWriter{logger: httpLogger, level: zapcore.ErrorLevel}

	// Setup router
	router := gin.New()

	server := &Server{
		router: router,
		logger: logger.Named("http"),
		server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
			Handler:      router,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			IdleTimeout:  config.IdleTimeout,
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(server)
	}

	return server
}

// RegisterRoutes registers routes with the router.
func (s *Server) RegisterRoutes(registrar RouterRegistrar) {
	registrar.Register(s.router)
}

// Start starts the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Stop gracefully shuts down the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
