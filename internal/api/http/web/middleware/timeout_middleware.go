package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
)

// Timeout middleware sets a maximum duration for request processing.
// This middleware complements server-level timeouts but focuses on request handling time.
// Note: This timeout should be shorter than the HTTP server's WriteTimeout.
// Parameters:
//   - timeout: Maximum allowed time for request processing
//
// Suggestions:
//   - If you expect certain API endpoints to take longer, consider setting a longer timeout for those endpoints
//   - For asynchronous tasks, they should return immediately and process in the background instead of relying on timeouts
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create timeout context
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Update request with timeout context
		c.Request = c.Request.WithContext(ctx)

		// Create completion channel
		done := make(chan struct{})

		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
			// Request completed before timeout
			return
		case <-ctx.Done():
			// Request timed out
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				c.AbortWithStatusJSON(http.StatusRequestTimeout, types.Error(types.CodeRequestTimeout, "Request timeout"))
			} else {
				c.Abort()
			}
		}
	}
}
