// Package middleware provides HTTP middleware for the Gin framework.
package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	appErrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/i18n"
	"github.com/Final-Year-Project-G22/backend/core/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger is a minimal interface for structured logging.
// This avoids importing internal/core which would create an import cycle.
// core.Logger satisfies this interface via structural typing.
type Logger interface {
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
}

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
// This is an enhanced version of gin.Recovery() with i18n support and structured logging.
func Recovery(logger Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := string(debug.Stack())
				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("ip", c.ClientIP()),
					zap.String("stack", stack),
				)

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

// RequestLogger returns a middleware that logs every HTTP request with structured logging.
func RequestLogger(logger Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		requestID, _ := c.Get(RequestIDKey)
		requestIDStr, _ := requestID.(string)

		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		fields := []zap.Field{
			zap.String("request_id", requestIDStr),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("route", route),
			zap.Int("status", status),
			zap.Int64("latency_ms", latency.Milliseconds()),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("response_size", c.Writer.Size()),
			zap.Int("error_count", len(c.Errors)),
		}

		if accountID := c.Value(contextkeys.AccountID); accountID != nil {
			if id := contextkeys.GetAccountID(accountID); id != contextkeys.NilUUID {
				fields = append(fields, zap.String("account_id", id.String()))
			}
		}
		if userID := c.Value(contextkeys.UserID); userID != nil {
			if id := contextkeys.GetUserID(userID); id != contextkeys.NilUUID {
				fields = append(fields, zap.String("user_id", id.String()))
			}
		}
		if sessionID := c.Value(contextkeys.SessionID); sessionID != nil {
			if id := contextkeys.GetSessionID(sessionID); id != contextkeys.NilUUID {
				fields = append(fields, zap.String("session_id", id.String()))
			}
		}
		if len(c.Errors) > 0 {
			fields = append(fields, zap.Error(c.Errors.Last().Err))
		}

		switch {
		case status >= 500:
			logger.Error("request", fields...)
		case status >= 400:
			logger.Warn("request", fields...)
		default:
			logger.Info("request", fields...)
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
