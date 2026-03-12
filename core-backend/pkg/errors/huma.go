package errors

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// HumaErrorDetail represents a single validation or sub-error detail.
type HumaErrorDetail struct {
	Message  string `json:"message" doc:"Human-readable error message"`
	Location string `json:"location,omitempty" doc:"Where the error occurred (e.g., 'body.email')"`
	Value    any    `json:"value,omitempty" doc:"The value that caused the error"`
}

// HumaError is a custom error model for Huma that matches our API conventions.
// It implements huma.StatusError to integrate with Huma's error handling.
type HumaError struct {
	status int
	Code   ErrorCode          `json:"code" doc:"Machine-readable error code"`
	Title  string             `json:"title" doc:"Short human-readable error title"`
	Detail string             `json:"detail" doc:"Detailed error message"`
	Errors []*HumaErrorDetail `json:"errors,omitempty" doc:"List of specific validation errors"`
}

// Error implements the error interface.
func (e *HumaError) Error() string {
	if e.Detail != "" {
		return e.Detail
	}
	return e.Title
}

// GetStatus implements huma.StatusError.
func (e *HumaError) GetStatus() int {
	return e.status
}

// statusToCode maps HTTP status codes to our ErrorCode constants.
func statusToCode(status int) ErrorCode {
	switch status {
	case http.StatusBadRequest:
		return ErrCodeBadRequest
	case http.StatusUnauthorized:
		return ErrCodeUnauthorized
	case http.StatusForbidden:
		return ErrCodeForbidden
	case http.StatusNotFound:
		return ErrCodeNotFound
	case http.StatusConflict:
		return ErrCodeConflict
	case http.StatusUnprocessableEntity:
		return ErrCodeValidation
	default:
		if status >= 400 && status < 500 {
			return ErrCodeBadRequest
		}
		return ErrCodeInternal
	}
}

// NewHumaError creates a new HumaError from status, message, and optional errors.
// This function is designed to replace huma.NewError.
func NewHumaError(status int, msg string, errs ...error) huma.StatusError {
	details := make([]*HumaErrorDetail, 0, len(errs))

	for _, err := range errs {
		if err == nil {
			continue
		}

		// Check if error implements huma.ErrorDetailer
		if detailer, ok := err.(huma.ErrorDetailer); ok {
			humaDetail := detailer.ErrorDetail()
			details = append(details, &HumaErrorDetail{
				Message:  humaDetail.Message,
				Location: humaDetail.Location,
				Value:    humaDetail.Value,
			})
		} else {
			details = append(details, &HumaErrorDetail{
				Message: err.Error(),
			})
		}
	}

	return &HumaError{
		status: status,
		Code:   statusToCode(status),
		Title:  http.StatusText(status),
		Detail: msg,
		Errors: details,
	}
}

// InitHumaErrorHandler sets up the custom error handler for Huma.
// Call this once during API initialization.
func InitHumaErrorHandler() {
	huma.NewError = NewHumaError
}
