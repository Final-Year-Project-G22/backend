// Package errors provides structured error types with i18n support.
package errors

import (
	"fmt"

	"github.com/Final-Year-Project-G22/backend/core/pkg/i18n"
)

// AppError represents an application error with localization support.
type AppError struct {
	Code    ErrorCode   // Unique error code (e.g., "validationError")
	Message string      // i18n key (e.g., "errors.userNotFound") or direct message
	Status  int         // HTTP status code
	Err     error       // Wrapped underlying error
	Details interface{} // Additional error details (for dev mode)
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Code, e.Err)
	}
	return fmt.Sprintf("%s", e.Code)
}

// Unwrap returns the underlying error for error wrapping.
func (e *AppError) Unwrap() error {
	return e.Err
}

// GetCode returns the error code.
func (e *AppError) GetCode() ErrorCode {
	return e.Code
}

// GetStatus returns the HTTP status code.
func (e *AppError) GetStatus() int {
	return e.Status
}

// GetMessage returns the resolved message for the given locale.
func (e *AppError) GetMessage(locale string) string {
	if e.Message == "" {
		return i18n.Resolve(string(e.Code), locale)
	}
	return i18n.Resolve(e.Message, locale)
}

// GetRawMessage returns the raw message key without resolution.
func (e *AppError) GetRawMessage() string {
	return e.Message
}

// WithDetails adds additional details to the error (typically for dev mode).
func (e *AppError) WithDetails(details interface{}) *AppError {
	e.Details = details
	return e
}

// NewError creates a new AppError with the given code, message, and status.
func NewError(code ErrorCode, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// NewErrorWithErr creates a new AppError with an underlying error.
func NewErrorWithErr(code ErrorCode, message string, status int, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
		Err:     err,
	}
}

// ValidationError creates a validation error.
func ValidationError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeValidation,
		Message: message,
		Status:  GetStatus(ErrCodeValidation),
	}
}

// NotFoundError creates a not found error.
// resource: the type of resource (e.g., "user")
// id: the ID that was not found
func NotFoundError(resource string, id interface{}) *AppError {
	key := fmt.Sprintf("errors.%sNotFound", resource)
	return &AppError{
		Code:    ErrCodeNotFound,
		Message: key,
		Status:  GetStatus(ErrCodeNotFound),
		Details: map[string]interface{}{
			"resource": resource,
			"id":       id,
		},
	}
}

// NotFoundErrorWithKey creates a not found error with a custom i18n key.
func NotFoundErrorWithKey(key string) *AppError {
	return &AppError{
		Code:    ErrCodeNotFound,
		Message: key,
		Status:  GetStatus(ErrCodeNotFound),
	}
}

// UnauthorizedError creates an unauthorized error.
func UnauthorizedError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeUnauthorized,
		Message: message,
		Status:  GetStatus(ErrCodeUnauthorized),
	}
}

// ForbiddenError creates a forbidden error.
func ForbiddenError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeForbidden,
		Message: message,
		Status:  GetStatus(ErrCodeForbidden),
	}
}

// ConflictError creates a conflict error.
func ConflictError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeConflict,
		Message: message,
		Status:  GetStatus(ErrCodeConflict),
	}
}

// AlreadyExistsError creates an "already exists" error.
func AlreadyExistsError(resource string, field string, value interface{}) *AppError {
	key := fmt.Sprintf("errors.%sAlreadyExists", resource)
	return &AppError{
		Code:    ErrCodeAlreadyExists,
		Message: key,
		Status:  GetStatus(ErrCodeAlreadyExists),
		Details: map[string]interface{}{
			"resource": resource,
			"field":    field,
			"value":    value,
		},
	}
}

// InternalError creates an internal server error.
func InternalError(message string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeInternal,
		Message: message,
		Status:  GetStatus(ErrCodeInternal),
		Err:     err,
	}
}

// BadRequestError creates a bad request error.
func BadRequestError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeBadRequest,
		Message: message,
		Status:  GetStatus(ErrCodeBadRequest),
	}
}

// InvalidInputError creates an invalid input error.
func InvalidInputError(field string, message string) *AppError {
	key := "errors.invalidInput"
	if message != "" {
		key = message
	}
	return &AppError{
		Code:    ErrCodeInvalidInput,
		Message: key,
		Status:  GetStatus(ErrCodeInvalidInput),
		Details: map[string]interface{}{
			"field": field,
		},
	}
}

// RequiredFieldError creates a required field error.
func RequiredFieldError(field string) *AppError {
	key := "errors.requiredField"
	return &AppError{
		Code:    ErrCodeValidation,
		Message: key,
		Status:  GetStatus(ErrCodeValidation),
		Details: map[string]interface{}{
			"field": field,
		},
	}
}
