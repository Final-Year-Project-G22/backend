// Package response provides standardized API response formatting.
// It integrates with the error handling system for consistent responses.
package response

import (
	"errors"

	appErrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// APIResponse is the standard response format for all API endpoints.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *Metadata   `json:"meta,omitempty"`
}

// APIError contains error information in the response.
type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Metadata contains pagination or additional response metadata.
type Metadata struct {
	Page       int   `json:"page,omitempty"`
	PageSize   int   `json:"page_size,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// Success sends a successful response with data.
//
// # Example
//
//	response.Success(c, gin.H{"user": "john"})
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, APIResponse{
		Success: true,
		Data:    data,
	})
}

// Created sends a 201 Created response.
//
// # Example
//
//	response.Created(c, newUser)
func Created(c *gin.Context, data interface{}) {
	c.JSON(201, APIResponse{
		Success: true,
		Data:    data,
	})
}

// NoContent sends a 204 No Content response.
func NoContent(c *gin.Context) {
	c.Status(204)
}

// Error sends an error response.
// It automatically resolves the i18n message based on the request's Accept-Language header.
//
// # Example
//
//	if err != nil {
//	    response.Error(c, err)
//	}
func Error(c *gin.Context, err error) {
	if err == nil {
		InternalServerError(c, "errors.unknownError")
		return
	}

	locale := getLocale(c)
	status := 500

	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {
		status = appErr.GetStatus()
		c.JSON(status, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    string(appErr.GetCode()),
				Message: appErr.GetMessage(locale),
			},
		})
		return
	}

	// Handle generic errors
	c.JSON(status, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "internalError",
			Message: i18n.Resolve("errors.internalError", locale),
		},
	})
}

// ErrorWithDetails sends an error response with additional details (for dev mode).
func ErrorWithDetails(c *gin.Context, err error, details interface{}) {
	if err == nil {
		InternalServerError(c, "errors.unknownError")
		return
	}

	locale := getLocale(c)
	status := 500

	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {
		status = appErr.GetStatus()
		c.JSON(status, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    string(appErr.GetCode()),
				Message: appErr.GetMessage(locale),
				Details: details,
			},
		})
		return
	}

	c.JSON(status, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "internalError",
			Message: i18n.Resolve("errors.internalError", locale),
			Details: details,
		},
	})
}

// BadRequest sends a 400 Bad Request error.
//
// # Example
//
//	response.BadRequest(c, "errors.invalidInput")
func BadRequest(c *gin.Context, message string) {
	locale := getLocale(c)
	c.JSON(400, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "badRequest",
			Message: i18n.Resolve(message, locale),
		},
	})
}

// Unauthorized sends a 401 Unauthorized error.
func Unauthorized(c *gin.Context, message string) {
	locale := getLocale(c)
	c.JSON(401, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "unauthorized",
			Message: i18n.Resolve(message, locale),
		},
	})
}

// Forbidden sends a 403 Forbidden error.
func Forbidden(c *gin.Context, message string) {
	locale := getLocale(c)
	c.JSON(403, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "forbidden",
			Message: i18n.Resolve(message, locale),
		},
	})
}

// NotFound sends a 404 Not Found error.
func NotFound(c *gin.Context, message string) {
	locale := getLocale(c)
	c.JSON(404, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "notFound",
			Message: i18n.Resolve(message, locale),
		},
	})
}

// Conflict sends a 409 Conflict error.
func Conflict(c *gin.Context, message string) {
	locale := getLocale(c)
	c.JSON(409, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "conflict",
			Message: i18n.Resolve(message, locale),
		},
	})
}

// InternalServerError sends a 500 Internal Server Error.
func InternalServerError(c *gin.Context, message string) {
	locale := getLocale(c)
	c.JSON(500, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "internalError",
			Message: i18n.Resolve(message, locale),
		},
	})
}

// ValidationError sends a 400 validation error response.
func ValidationError(c *gin.Context, validationErrs appErrors.FieldValidationErrors) {
	locale := getLocale(c)
	formatted := appErrors.FormatValidationErrors(validationErrs, locale)

	c.JSON(400, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "validationError",
			Message: i18n.Resolve("errors.validationError", locale),
			Details: formatted,
		},
	})
}

// Paginated sends a paginated response.
//
// # Example
//
//	response.Paginated(c, users, total, page, pageSize)
func Paginated(c *gin.Context, data interface{}, total int64, page, pageSize int) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(200, APIResponse{
		Success: true,
		Data:    data,
		Meta: &Metadata{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// getLocale extracts the locale from the request's Accept-Language header.
func getLocale(c *gin.Context) string {
	locale := c.GetHeader("Accept-Language")
	if locale == "" {
		return i18n.GetDefaultLocale()
	}
	return locale
}
