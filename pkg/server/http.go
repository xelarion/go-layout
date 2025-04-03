// Package server provides server implementations for the application.
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HTTPConfig holds configuration for the HTTP server.
type HTTPConfig struct {
	Host string
	Port int
	Mode string // debug, release, test
	// ReadTimeout is the maximum duration for reading the entire request
	ReadTimeout time.Duration
	// WriteTimeout is the maximum duration before timing out writes of the response
	WriteTimeout time.Duration
	// IdleTimeout is the maximum amount of time to wait for the next request
	IdleTimeout time.Duration
}

// HTTPServer represents an HTTP server.
type HTTPServer struct {
	server *http.Server
	router *gin.Engine
	logger *zap.Logger
	config *HTTPConfig
}

// NewHTTPServer creates a new HTTP server instance.
func NewHTTPServer(config *HTTPConfig, logger *zap.Logger) *HTTPServer {
	// Set Gin mode
	gin.SetMode(config.Mode)

	// Setup router
	router := gin.New()

	return &HTTPServer{
		router: router,
		logger: logger,
		config: config,
		server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
			Handler:      router,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			IdleTimeout:  config.IdleTimeout,
		},
	}
}

// Router returns the Gin router instance.
func (s *HTTPServer) Router() *gin.Engine {
	return s.router
}

// Start starts the HTTP server.
func (s *HTTPServer) Start() error {
	s.logger.Info("Starting HTTP server", zap.String("addr", s.server.Addr))
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown gracefully shuts down the HTTP server.
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server")
	return s.server.Shutdown(ctx)
}
