package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS middleware configures Cross-Origin Resource Sharing (CORS) policies.
// This allows frontend applications (like Vue.js) to make API requests to this backend.
// Uses production-ready best practices for security and performance.
// Parameters:
//   - allowOrigins: List of allowed origins for CORS
func CORS(allowOrigins []string) gin.HandlerFunc {
	// Configure CORS with secure production defaults
	corsConfig := cors.Config{
		AllowOrigins: allowOrigins,
		// Allow standard HTTP methods needed for RESTful APIs
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		// Allow essential headers needed for modern web applications
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},
		// Expose only necessary headers
		ExposeHeaders: []string{"Content-Length", "Content-Type", "Content-Disposition"},
		// Allow credentials for authenticated requests
		AllowCredentials: true,
		// Cache preflight requests for 1 hour
		MaxAge: time.Hour,
	}

	// Set AllowAllOrigins flag if wildcard origin is specified
	if len(corsConfig.AllowOrigins) == 1 && corsConfig.AllowOrigins[0] == "*" {
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowOrigins = nil
	}

	return cors.New(corsConfig)
}
