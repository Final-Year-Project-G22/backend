package middleware

import (
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/token"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/danielgtaylor/huma/v2"
)

// AuthMiddleware creates a Huma router-agnostic middleware that validates
// Bearer tokens and injects session claims into the context.
//
// This middleware:
// - Extracts the Bearer token from the Authorization header
// - Validates the JWT using the TokenService
// - Injects SessionID and Email into the Huma context via huma.WithValue
// - Returns 401 Unauthorized if the token is missing or invalid
//
// Apply this middleware only to protected operations via huma.Operation{Middlewares: ...}
func AuthMiddleware(api huma.API, tokenService token.TokenService, authService service.AuthService) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		// Get Authorization header
		authHeader := ctx.Header("Authorization")
		if authHeader == "" {
			_ = huma.WriteErr(api, ctx, 401, "missing authorization header")
			return
		}

		// Check Bearer prefix
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			_ = huma.WriteErr(api, ctx, 401, "invalid authorization header format")
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)
		if tokenString == "" {
			_ = huma.WriteErr(api, ctx, 401, "empty token")
			return
		}

		// Validate token
		claims, err := tokenService.ValidateAccessToken(ctx.Context(), tokenString)
		if err != nil {
			_ = huma.WriteErr(api, ctx, 401, "invalid or expired token")
			return
		}

		output, err := authService.ValidateAccessSession(ctx.Context(), claims.SessionID, false)

		if err != nil {
			if appErr, ok := err.(*apperrors.AppError); ok {
				_ = huma.WriteErr(api, ctx, appErr.GetStatus(), appErr.GetMessage("en"))
				return
			}

			_ = huma.WriteErr(api, ctx, 401, "session is no longer active")
			return
		}

		// Inject claims into context using huma.WithValue
		ctx = huma.WithValue(ctx, contextkeys.SessionID, claims.SessionID)
		ctx = huma.WithValue(ctx, contextkeys.Email, claims.Email)
		ctx = huma.WithValue(ctx, contextkeys.AccountID, output.AccountID)
		ctx = huma.WithValue(ctx, contextkeys.UserID, output.UserID)

		// Continue to next handler with enriched context
		next(ctx)
	}
}

func AccountStatusMiddleware(api huma.API, authService service.AuthService) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		sessionID := contextkeys.GetSessionID(
			ctx.Context().Value(contextkeys.SessionID),
		)
		if sessionID == contextkeys.NilUUID {
			_ = huma.WriteErr(api, ctx, 401, "missing session context")
			return
		}
		_, err := authService.ValidateAccessSession(ctx.Context(), sessionID, true)
		if err != nil {
			if appErr, ok := err.(*apperrors.AppError); ok {
				_ = huma.WriteErr(api, ctx, appErr.GetStatus(), appErr.GetMessage("en"))
				return
			}
			_ = huma.WriteErr(api, ctx, 403, "account not active")
			return
		}
		next(ctx)
	}
}
