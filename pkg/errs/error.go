// Package errs provides error handling utilities for the application.
package errs

import (
	"errors"
	"fmt"
	"maps"
	"runtime"
	"strings"
)

// ErrorType represents the type of an error
type ErrorType int

const (
	// InternalError represents errors from system components like database, cache, etc.
	InternalError ErrorType = iota
	// BusinessError represents errors related to business logic violations
	BusinessError
	// ValidationError represents user input validation errors (a specialized business error)
	ValidationError
)

// Common error reasons as slugs
const (
	// Default reason

	ReasonUnknown = "UNKNOWN" // Unknown or unspecified reason

	// Data related errors

	ReasonNotFound  = "NOT_FOUND" // Record not found
	ReasonDuplicate = "DUPLICATE" // Duplicate record

	// Permission related errors

	ReasonUnauthorized = "UNAUTHORIZED"  // Not authenticated (not logged in)
	ReasonUserDisabled = "USER_DISABLED" // User account is disabled
	ReasonForbidden    = "FORBIDDEN"     // Access denied (no permission)

	// Client errors

	ReasonBadRequest   = "BAD_REQUEST"   // Invalid request parameters or format
	ReasonInvalidState = "INVALID_STATE" // Invalid state, business rule conflict

	// Server errors

	ReasonInternalError = "INTERNAL_ERROR" // Internal error (generic)
)

// Error represents a custom error with context information
type Error struct {
	errType ErrorType      // Type of error (internal, business, validation)
	message string         // User-friendly error message
	reason  string         // Machine-readable reason code (slug)
	err     error          // Original error
	stack   string         // Stack trace where the error occurred
	meta    map[string]any // Additional metadata for logging and debugging
}

// Error returns the string representation of the error
func (e *Error) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %v", e.message, e.err)
	}
	return e.message
}

// Unwrap implements the errors.Unwrap interface for compatibility with errors.Is/As
func (e *Error) Unwrap() error {
	return e.err
}

// Is implements the interface for errors.Is
func (e *Error) Is(target error) bool {
	// Check if target is the same type and shares the same reason/type
	var t *Error
	if errors.As(target, &t) {
		sameType := e.errType == t.errType
		sameReason := e.reason != "" && t.reason != "" && e.reason == t.reason
		return sameType || sameReason
	}
	return false
}

// Type returns the error type
func (e *Error) Type() ErrorType {
	return e.errType
}

// Message returns the error message
func (e *Error) Message() string {
	return e.message
}

// Reason returns the error reason
func (e *Error) Reason() string {
	return e.reason
}

// Stack returns the error stack trace
func (e *Error) Stack() string {
	return e.stack
}

// Meta returns the error metadata
func (e *Error) Meta() map[string]any {
	return e.meta
}

// WithMeta adds a key-value pair to the error's metadata
func (e *Error) WithMeta(key string, value any) *Error {
	if e.meta == nil {
		e.meta = make(map[string]any)
	}
	e.meta[key] = value
	return e
}

// WithMetaMap adds a map of key-value pairs to the error's metadata
func (e *Error) WithMetaMap(data map[string]any) *Error {
	if e.meta == nil {
		e.meta = make(map[string]any)
	}
	maps.Copy(e.meta, data)
	return e
}

// WithReason sets the reason for the error
func (e *Error) WithReason(reason string) *Error {
	e.reason = reason
	return e
}

// GetMeta gets metadata from the error
func GetMeta(err error) map[string]any {
	var e *Error
	if errors.As(err, &e) {
		return e.meta
	}
	return nil
}

// GetReason gets the reason code from the error if available
func GetReason(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.reason
	}
	return ""
}

// GetMessage gets the message from the error if available
func GetMessage(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.message
	}
	return err.Error()
}

// GetStack gets the stack trace from the error if available
func GetStack(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.stack
	}
	return ""
}

// IsInternal checks if the error is an internal error
func IsInternal(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.errType == InternalError
	}
	return false
}

// IsBusiness checks if the error is a business error
func IsBusiness(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.errType == BusinessError
	}
	return false
}

// IsValidation checks if the error is a validation error
func IsValidation(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.errType == ValidationError
	}
	return false
}

// IsReason checks if the error has a specific reason code
func IsReason(err error, reason string) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.reason == reason
	}
	return false
}

// NewInternal creates a new internal (system) error.
func NewInternal(message string) *Error {
	return &Error{
		errType: InternalError,
		message: message,
		reason:  ReasonInternalError, // Default reason for internal errors
		stack:   getStackTrace(1),
	}
}

// NewBusiness creates a new business logic error.
func NewBusiness(message string) *Error {
	return &Error{
		errType: BusinessError,
		message: message,
		reason:  ReasonUnknown, // Default to unknown reason for business errors
		stack:   getStackTrace(1),
	}
}

// NewValidation creates a new validation error.
func NewValidation(message string) *Error {
	return &Error{
		errType: ValidationError,
		message: message,
		reason:  ReasonBadRequest, // Default reason for validation errors
		stack:   getStackTrace(1),
	}
}

// WrapInternal wraps an existing error as an internal error.
func WrapInternal(err error, message string) *Error {
	return &Error{
		errType: InternalError,
		message: message,
		reason:  ReasonInternalError, // Default reason for internal errors
		err:     err,
		stack:   getStackTrace(1),
	}
}

// WrapBusiness wraps an error as a business error
func WrapBusiness(err error, message string) *Error {
	return &Error{
		errType: BusinessError,
		message: message,
		err:     err,
		stack:   getStackTrace(1),
	}
}

// WrapValidation wraps an error as a validation error
func WrapValidation(err error, message string) *Error {
	return &Error{
		errType: ValidationError,
		message: message,
		err:     err,
		stack:   getStackTrace(1),
	}
}

// Unwrap implements the errors.Unwrap interface for compatibility with errors.Is/As
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// getStackTrace returns a formatted stack trace
func getStackTrace(skip int) string {
	stackBuf := make([]uintptr, 20) // Reduced from 50 to 20 for more compact traces
	length := runtime.Callers(skip+2, stackBuf)
	stack := stackBuf[:length]

	var traceBuf strings.Builder
	traceBuf.Grow(500)
	frames := runtime.CallersFrames(stack)
	for {
		frame, more := frames.Next()
		// Skip runtime and errs package frames for cleaner traces
		if strings.Contains(frame.File, "runtime/") || strings.Contains(frame.File, "/errs/") {
			if more {
				continue
			}
			break
		}
		fmt.Fprintf(&traceBuf, "\n\t%s:%d %s", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}
	return traceBuf.String()
}
