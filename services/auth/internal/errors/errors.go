package errors

import (
	"encoding/json"
	"net/http"
)

// ErrorCode represents a specific error type
type ErrorCode string

const (
	ErrorCodeValidation      ErrorCode = "VALIDATION_ERROR"
	ErrorCodeUnauthorized    ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden       ErrorCode = "FORBIDDEN"
	ErrorCodeNotFound        ErrorCode = "NOT_FOUND"
	ErrorCodeConflict        ErrorCode = "CONFLICT"
	ErrorCodeInternal        ErrorCode = "INTERNAL_ERROR"
	ErrorCodeBadRequest      ErrorCode = "BAD_REQUEST"
	ErrorCodeEmailNotVerified ErrorCode = "EMAIL_NOT_VERIFIED"
)

// AppError represents an application error
type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details []string  `json:"details,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return e.Message
}

// HTTPStatus returns the appropriate HTTP status code for the error
func (e *AppError) HTTPStatus() int {
	switch e.Code {
	case ErrorCodeValidation, ErrorCodeBadRequest:
		return http.StatusBadRequest
	case ErrorCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrorCodeForbidden:
		return http.StatusForbidden
	case ErrorCodeNotFound:
		return http.StatusNotFound
	case ErrorCodeConflict:
		return http.StatusConflict
	case ErrorCodeEmailNotVerified:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

// ErrorResponse is the standard error response format
type ErrorResponse struct {
	Error struct {
		Code    ErrorCode `json:"code"`
		Message string    `json:"message"`
		Details []string  `json:"details,omitempty"`
	} `json:"error"`
}

// NewError creates a new AppError
func NewError(code ErrorCode, message string, details ...string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// ToResponse converts an AppError to ErrorResponse
func (e *AppError) ToResponse() ErrorResponse {
	var resp ErrorResponse
	resp.Error.Code = e.Code
	resp.Error.Message = e.Message
	resp.Error.Details = e.Details
	return resp
}

// WriteError writes an error response to the HTTP response
func WriteError(w http.ResponseWriter, err *AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatus())
	json.NewEncoder(w).Encode(err.ToResponse())
}

