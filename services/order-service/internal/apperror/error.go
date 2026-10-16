package apperror

import (
	"errors"
	"fmt"
)

// ErrorCode represents a standardized domain or application error code.
type ErrorCode string

const (
	CodeInternal     ErrorCode = "INTERNAL_SERVER_ERROR"
	CodeInvalidInput ErrorCode = "INVALID_INPUT"
	CodeNotFound     ErrorCode = "RESOURCE_NOT_FOUND"
	CodeConflict     ErrorCode = "CONFLICT"
)

// AppError is the unified application error structure.
type AppError struct {
	Code    ErrorCode         `json:"code"`
	Message string            `json:"message"`
	Err     error             `json:"-"`
	Details map[string]string `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Convenience constructors

func NewInternal(message string, err error) *AppError {
	return &AppError{
		Code:    CodeInternal,
		Message: message,
		Err:     err,
	}
}

func NewInvalidInput(message string, details map[string]string) *AppError {
	return &AppError{
		Code:    CodeInvalidInput,
		Message: message,
		Details: details,
	}
}

func NewNotFound(message string) *AppError {
	return &AppError{
		Code:    CodeNotFound,
		Message: message,
	}
}

func NewConflict(message string, err error) *AppError {
	return &AppError{
		Code:    CodeConflict,
		Message: message,
		Err:     err,
	}
}

// AsAppError extracts *AppError from an error chain if present.
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
