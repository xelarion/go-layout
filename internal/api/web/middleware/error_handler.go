// Package middleware contains HTTP middleware components.
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/web/types"
	"github.com/xelarion/go-layout/pkg/errs"
)

// ErrorHandler adds centralized error handling and logging for the API layer.
func ErrorHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process the request
		c.Next()

		// Check if there are any errors
		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		var statusCode int
		var errCode int
		var errMsg string

		// Prepare fields for logging
		fields := []zap.Field{
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("client_ip", c.ClientIP()),
		}

		// Add metadata if available
		meta := errs.GetMeta(err)
		if meta != nil {
			for k, v := range meta {
				fields = append(fields, zap.Any(k, v))
			}
		}

		// Add reason if available
		reason := errs.GetReason(err)
		if reason != "" {
			fields = append(fields, zap.String("reason", reason))
		}

		// Default error handling
		statusCode = http.StatusInternalServerError
		errCode = types.CodeInternalError
		// Get a user-friendly message
		errMsg = getMessage(err)

		// Determine the appropriate status code and error code based on error type
		if errs.IsInternal(err) {
			statusCode = http.StatusInternalServerError
			errCode = types.CodeInternalError
			
			// Add stack trace for internal errors to aid debugging
			stack := errs.GetStack(err)
			if stack != "" {
				fields = append(fields, zap.String("stack_trace", stack))
			}
			
			logger.Error("Internal server error", fields...)
		} else if errs.IsBusiness(err) {
			statusCode = http.StatusBadRequest
			errCode = types.CodeBadRequest

			// Map common error reasons to their respective status codes
			switch reason {
			case errs.ReasonNotFound:
				statusCode = http.StatusNotFound
				errCode = types.CodeNotFound
			case errs.ReasonUnauthorized:
				statusCode = http.StatusUnauthorized
				errCode = types.CodeUnauthorized
			case errs.ReasonForbidden:
				statusCode = http.StatusForbidden
				errCode = types.CodeForbidden
			case errs.ReasonDuplicate:
				statusCode = http.StatusConflict
				errCode = types.CodeDuplicate
			case errs.ReasonBadRequest:
				// For validation errors, we stay with 400
				errCode = types.CodeValidation
				logger.Debug("Validation error", fields...)
			default:
				// For any other business error, use the same default
				errCode = types.CodeBadRequest
				logger.Info("Business error", fields...)
			}

			logger.Info("Business logic error", fields...)
		} else if errs.IsValidation(err) {
			statusCode = http.StatusBadRequest
			errCode = types.CodeValidation
			logger.Info("Validation error", fields...)
		} else {
			// Unknown error type
			logger.Error("Unhandled error", append(fields, zap.Error(err))...)
		}

		// Send the response using the standard types.Response structure
		c.JSON(statusCode, types.Error(errCode, errMsg))
		c.Abort()
	}
}

// getMessage gets a user-appropriate error message
func getMessage(err error) string {
	message := errs.GetMessage(err)
	
	// For internal errors, we don't want to leak implementation details
	if errs.IsInternal(err) {
		return "Internal server error"
	}
	
	return message
}
