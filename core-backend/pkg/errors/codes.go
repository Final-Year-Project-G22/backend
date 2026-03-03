// Package errors provides structured error types for the application.
// It integrates with i18n for localized error messages.
package errors

// ErrorCode represents a unique error code for categorization.
// Using camelCase as per project convention.
type ErrorCode string

// Common error codes used throughout the application.
const (
	ErrCodeValidation    ErrorCode = "validationError"
	ErrCodeNotFound      ErrorCode = "notFound"
	ErrCodeUnauthorized  ErrorCode = "unauthorized"
	ErrCodeForbidden     ErrorCode = "forbidden"
	ErrCodeConflict      ErrorCode = "conflict"
	ErrCodeInternal      ErrorCode = "internalError"
	ErrCodeBadRequest    ErrorCode = "badRequest"
	ErrCodeAlreadyExists ErrorCode = "alreadyExists"
	ErrCodeInvalidInput  ErrorCode = "invalidInput"
)

// HTTP status mapping for error codes.
var codeToStatus = map[ErrorCode]int{
	ErrCodeValidation:    400,
	ErrCodeNotFound:      404,
	ErrCodeUnauthorized:  401,
	ErrCodeForbidden:     403,
	ErrCodeConflict:      409,
	ErrCodeInternal:      500,
	ErrCodeBadRequest:    400,
	ErrCodeAlreadyExists: 409,
	ErrCodeInvalidInput:  400,
}

// GetStatus returns the HTTP status code for an error code.
func GetStatus(code ErrorCode) int {
	if status, ok := codeToStatus[code]; ok {
		return status
	}
	return 500
}
