package types

// Common business status codes
const (
	// Success codes (0-999)
	CodeSuccess = 0 // General success

	// Client error codes (1000-1499)
	CodeBadRequest     = 1000 // Invalid request parameters or format
	CodeUnauthorized   = 1001 // Authentication required (not logged in)
	CodeForbidden      = 1002 // Permission denied (no access rights)
	CodeValidation     = 1003 // Input validation failed
	CodeNotFound       = 1004 // Resource not found
	CodeDuplicate      = 1005 // Resource already exists (duplicate)
	CodeRequestTimeout = 1006 // Request timeout

	// Business-specific error codes (2000-4999)
	CodeInvalidState = 2000 // Invalid business state or rule violation

	// Server/System error codes (5000+)
	CodeInternalError = 5000 // Server internal error
)

// Response represents the standard API response structure.
type Response struct {
	Code    int    `json:"code"`           // Business status code
	Message string `json:"message"`        // Response message
	Data    any    `json:"data,omitempty"` // Data payload
	Meta    any    `json:"meta,omitempty"` // Additional metadata
}

// NewResponse creates a new response with the given parameters.
func NewResponse(code int, message string, data any, meta any) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Data:    data,
		Meta:    meta,
	}
}

// Error creates an error response with no data.
func Error(code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
	}
}

// Success creates a success response with data.
func Success(data any) *Response {
	return &Response{
		Code:    CodeSuccess,
		Message: "",
		Data:    data,
	}
}

// WithMessage adds a message to the response.
func (r *Response) WithMessage(message string) *Response {
	r.Message = message
	return r
}

// WithMeta adds metadata to the response.
func (r *Response) WithMeta(meta any) *Response {
	r.Meta = meta
	return r
}
