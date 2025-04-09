// Package middleware contains Web API middleware components.
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	"github.com/xelarion/go-layout/pkg/errs"
)

// Error returns a middleware that handles errors and logs them using zap
// It should be registered at the router level to catch all errors.
func Error(logger *zap.Logger) gin.HandlerFunc {
	logger = logger.WithOptions(
		zap.WithCaller(false),
		zap.AddStacktrace(zap.FatalLevel),
	)

	return func(c *gin.Context) {
		// Process the request
		c.Next()

		// Check if there are any errors
		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		// Prepare fields for logging
		fields := []zap.Field{
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("client_ip", c.ClientIP()),
			zap.Error(err), // Always include the original error
		}

		// Add metadata if available
		meta := errs.GetMeta(err)
		if meta != nil && len(meta) > 0 {
			fields = append(fields, zap.Any("metadata", meta))
		}

		// Add reason if available
		reason := errs.GetReason(err)
		if reason != "" {
			fields = append(fields, zap.String("reason", reason))
		}

		var httpStatus int
		var respCode int
		respMessage := getErrorMessage(err)

		// Determine the appropriate status code and error code based on error type
		if errs.IsInternal(err) {
			httpStatus = http.StatusInternalServerError
			respCode = types.CodeInternalError

			// Add stack trace for internal errors to aid debugging
			stack := errs.GetStack(err)
			if stack != "" {
				fields = append(fields, zap.String("stack_trace", stack))
			}

			logger.Error("Internal server error", fields...)
		} else if errs.IsBusiness(err) {
			// Map common error reasons to their respective status codes
			switch reason {
			case errs.ReasonNotFound:
				httpStatus = http.StatusNotFound
				respCode = types.CodeNotFound
			case errs.ReasonDuplicate:
				httpStatus = http.StatusConflict
				respCode = types.CodeDuplicate
			case errs.ReasonUnauthorized:
				httpStatus = http.StatusUnauthorized
				respCode = types.CodeUnauthorized
			case errs.ReasonForbidden:
				httpStatus = http.StatusForbidden
				respCode = types.CodeForbidden
			case errs.ReasonBadRequest:
				httpStatus = http.StatusBadRequest
				respCode = types.CodeValidation
			case errs.ReasonInvalidState:
				httpStatus = http.StatusConflict
				respCode = types.CodeInvalidState
			default:
				// For any other business error, use the same default
				httpStatus = http.StatusBadRequest
				respCode = types.CodeBadRequest
			}

			logger.Info("Business logic error", fields...)
		} else if errs.IsValidation(err) {
			httpStatus = http.StatusBadRequest
			respCode = types.CodeValidation
			logger.Info("Validation error", fields...)
		} else {
			// Unknown error type
			httpStatus = http.StatusInternalServerError
			respCode = types.CodeInternalError
			logger.Error("Unhandled error", fields...)
		}

		// Send the response using the standard types.Response structure
		c.JSON(httpStatus, types.Error(respCode, respMessage))
		c.Abort()
	}
}

// getErrorMessage gets a user-appropriate error message
func getErrorMessage(err error) string {
	message := errs.GetMessage(err)

	// For internal errors, we don't want to leak implementation details
	if errs.IsInternal(err) {
		return "Internal server error"
	}

	return message
}
