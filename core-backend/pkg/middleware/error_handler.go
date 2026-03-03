// Package middleware provides HTTP middleware for the Gin framework.
package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	appErrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/i18n"
	"github.com/Final-Year-Project-G22/backend/core/pkg/response"
	"github.com/gin-gonic/gin"
)

const (
	// PanicRecoveryKey is the context key for panic information.
	PanicRecoveryKey = "panic_recovery"
)

// ErrorHandler returns a Gin middleware that handles errors and panics globally.
//
// # Features
//
//   - Recovers from panics and returns a 500 error
//   - Converts AppError to proper HTTP responses
//   - Logs errors with stack traces in development mode
//   - Shows detailed error information only in debug mode
//
// # Usage
//
//	router := gin.New()
//	router.Use(middleware.ErrorHandler())
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			handleError(c, err.Err)
		}
	}
}

// Recovery returns a Gin middleware that recovers from panics.
// This is a enhanced version of gin.Recovery() with i18n support.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := string(debug.Stack())
				log.Printf("[PANIC] %v\n%s", err, stack)

				c.Set(PanicRecoveryKey, err)
				c.Abort()

				locale := getLocale(c)
				message := i18n.Resolve("errors.internalError", locale)

				isDebug := gin.Mode() == gin.DebugMode
				if isDebug {
					response.ErrorWithDetails(c, fmt.Errorf("%v", err), map[string]interface{}{
						"panic":   err,
						"stack":   stack,
						"message": message,
					})
					return
				}

				response.InternalServerError(c, "errors.internalError")
			}
		}()
		c.Next()
	}
}

// handleError handles an error and sends an appropriate response.
func handleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	isDebug := gin.Mode() == gin.DebugMode

	// Check if it's an AppError
	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {
		if isDebug && appErr.Err != nil {
			response.ErrorWithDetails(c, err, map[string]interface{}{
				"stack": appErr.Err.Error(),
			})
			return
		}
		response.Error(c, err)
		return
	}

	// Handle generic errors
	if isDebug {
		response.ErrorWithDetails(c, err, map[string]string{
			"stack": fmt.Sprintf("%+v", err),
		})
		return
	}

	// Hide error details in production
	response.InternalServerError(c, "errors.internalError")
}

// getLocale extracts the locale from the request's Accept-Language header.
func getLocale(c *gin.Context) string {
	locale := c.GetHeader("Accept-Language")
	if locale == "" {
		return i18n.GetDefaultLocale()
	}
	return locale
}

// RequestLogger returns a middleware that logs HTTP requests.
// Enhanced version with i18n-aware error logging.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Log errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			log.Printf("[ERROR] %s - %s - %v", c.Request.Method, c.Request.URL.Path, err.Err)
		}
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing.
func CORSMiddleware(allowedOrigins ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, o := range allowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			if origin != "" {
				c.Header("Access-Control-Allow-Origin", origin)
			}
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
