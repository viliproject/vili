package errors

import (
	"errors"
	"net/http"
)

// ErrorResponse is the json wrapper for API error responses
type ErrorResponse struct {
	Status    int                 `json:"-"`
	ErrorType string              `json:"type"`
	Message   string              `json:"message"`
	Params    map[string][]string `json:"params,omitempty"`
}

func (e *ErrorResponse) Error() string {
	return e.Message
}

// New returns a new error
func New(err string) error {
	return errors.New(err)
}

// InternalServerError returns an internal server error response
func InternalServerError() *ErrorResponse {
	return &ErrorResponse{
		Status:    http.StatusInternalServerError,
		ErrorType: "internal_server_error",
		Message:   "Internal server error",
	}
}

// UnauthorizedError returns an unauthorized error response
func UnauthorizedError(message string) *ErrorResponse {
	if message == "" {
		message = "Unauthorized"
	}
	return &ErrorResponse{
		Status:    http.StatusUnauthorized,
		ErrorType: "unauthorized",
		Message:   message,
	}
}

// ForbiddenError returns a forbidden error response
func ForbiddenError(message string) *ErrorResponse {
	if message == "" {
		message = "Forbidden"
	}
	return &ErrorResponse{
		Status:    http.StatusForbidden,
		ErrorType: "forbidden",
		Message:   message,
	}
}

// BadRequestError returns a bad request error response
func BadRequestError(message string) *ErrorResponse {
	if message == "" {
		message = "Bad request"
	}
	return &ErrorResponse{
		Status:    http.StatusBadRequest,
		ErrorType: "bad_request",
		Message:   message,
	}
}

// NotFoundError returns a not found error response
func NotFoundError(message string) *ErrorResponse {
	if message == "" {
		message = "Not found"
	}
	return &ErrorResponse{
		Status:    http.StatusNotFound,
		ErrorType: "not_found",
		Message:   message,
	}
}

// ValidationError returns a validation error response
func ValidationError(message string, params map[string][]string) *ErrorResponse {
	if message == "" {
		message = "Validation error"
	}
	return &ErrorResponse{
		Status:    http.StatusBadRequest,
		ErrorType: "validation_error",
		Message:   message,
		Params:    params,
	}
}

// ConflictError returns a not found error response
func ConflictError(message string) *ErrorResponse {
	if message == "" {
		message = "Conflict"
	}
	return &ErrorResponse{
		Status:    http.StatusConflict,
		ErrorType: "conflict",
		Message:   message,
	}
}

// RequestTooLargeError returns a request too large error response
func RequestTooLargeError(message string) *ErrorResponse {
	if message == "" {
		message = "Request entity too large"
	}
	return &ErrorResponse{
		Status:    http.StatusRequestEntityTooLarge,
		ErrorType: "request_entity_too_large",
		Message:   message,
	}
}
