package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
)

// zapWriter implements the io.Writer interface to redirect gin's recovery logs to zap
type zapWriter struct {
	logger *zap.Logger
}

// Write satisfies the io.Writer interface and forwards the message to zap
func (w *zapWriter) Write(p []byte) (n int, err error) {
	// Trim any trailing newlines
	w.logger.Error("Recovery from panic"+string(p),
		zap.Error(err),
		zap.String("time", time.Now().Format(time.RFC3339)),
	)
	return len(p), nil
}

// Recovery returns a middleware that recovers from panics and logs using our zap logger
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	// Create a writer that will redirect output to zap
	writer := &zapWriter{logger: logger.WithOptions(
		zap.WithCaller(false),
		zap.AddStacktrace(zap.FatalLevel),
	)}

	// Use gin's built-in recovery, but with our custom writer
	ginRecovery := gin.RecoveryWithWriter(writer, func(c *gin.Context, err any) {
		// Send a 500 response with our standard error format
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			types.Error(types.CodeInternalError, "Internal server error"))
	})

	return ginRecovery
}
