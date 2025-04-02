// Package middleware contains HTTP middleware functions for the application.
package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/xelarion/go-layout/internal/api/web/types"
)

// Timeout middleware sets a timeout for the request and aborts it if it takes too long.
// This helps prevent long-running requests from consuming server resources indefinitely.
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Update the request with the timeout context
		c.Request = c.Request.WithContext(ctx)

		// Create a channel to signal completion
		done := make(chan struct{})

		go func() {
			// Process the request
			c.Next()
			// Signal completion
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
