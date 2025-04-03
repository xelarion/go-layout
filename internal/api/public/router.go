// Package public contains Public API handlers and routers for external clients.
package public

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Router handles all routes for the Public API.
type Router struct {
	Engine *gin.Engine
	logger *zap.Logger
}

// NewRouter creates a new Public API router.
func NewRouter(engine *gin.Engine, logger *zap.Logger) *Router {
	return &Router{
		Engine: engine,
		logger: logger.Named("public_router"),
	}
}

// SetupRoutes configures all routes for the Public API.
func (r *Router) SetupRoutes() {
}
