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

// Unauthorized returns an unauthorized error response
func Unauthorized(message string) *ErrorResponse {
	if message == "" {
		message = "Unauthorized"
	}
	return &ErrorResponse{
		Status:    http.StatusUnauthorized,
		ErrorType: "unauthorized",
		Message:   message,
	}
}

// Forbidden returns a forbidden error response
func Forbidden(message string) *ErrorResponse {
	if message == "" {
		message = "Forbidden"
	}
	return &ErrorResponse{
		Status:    http.StatusForbidden,
		ErrorType: "forbidden",
		Message:   message,
	}
}

// BadRequest returns a bad request error response
func BadRequest(message string) *ErrorResponse {
	if message == "" {
		message = "Bad request"
	}
	return &ErrorResponse{
		Status:    http.StatusBadRequest,
		ErrorType: "bad_request",
		Message:   message,
	}
}

// NotFound returns a not found error response
func NotFound(message string) *ErrorResponse {
	if message == "" {
		message = "Not found"
	}
	return &ErrorResponse{
		Status:    http.StatusNotFound,
		ErrorType: "not_found",
		Message:   message,
	}
}

// MethodNotAllowed returns a method not allowed error response
func MethodNotAllowed(message string) *ErrorResponse {
	if message == "" {
		message = "Method not allowed"
	}
	return &ErrorResponse{
		Status:    http.StatusMethodNotAllowed,
		ErrorType: "method_not_allowed",
		Message:   message,
	}
}

// Validation returns a validation error response
func Validation(message string, params map[string][]string) *ErrorResponse {
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

// Conflict returns a not found error response
func Conflict(message string) *ErrorResponse {
	if message == "" {
		message = "Conflict"
	}
	return &ErrorResponse{
		Status:    http.StatusConflict,
		ErrorType: "conflict",
		Message:   message,
	}
}
