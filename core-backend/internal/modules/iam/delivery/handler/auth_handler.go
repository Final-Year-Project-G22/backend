package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/token"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/i18n"
)

const refreshCookiePath = "/api/v1/auth/refresh"

type AuthHandler struct {
	authService  service.AuthService
	tokenService token.TokenService
	cfg          *core.Config
}

func NewAuthHandler(authService service.AuthService, tokenService token.TokenService, cfg *core.Config) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		tokenService: tokenService,
		cfg:          cfg,
	}
}

// HandleRegister handles POST /api/v1/auth/register
func (h *AuthHandler) HandleRegister(ctx context.Context, input *dto.RegisterInput) (*dto.RegisterOutput, error) {
	result, err := h.authService.Register(ctx, service.RegisterInput{
		Email:     input.Body.Email,
		Password:  input.Body.Password,
		FirstName: input.Body.FirstName,
		LastName:  input.Body.LastName,
		// TODO: Extract UserAgent and IPAddress from request headers if needed
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.RegisterOutput{
		SetCookie: h.createRefreshTokenCookie(result.RefreshToken),
		Body: dto.RegisterResponseBody{
			AccessToken: result.AccessToken,
			ExpiresAt:   result.ExpiresAt,
			User:        toUserDTO(result.User),
			Account:     toAccountDTO(result.Account),
		},
	}, nil
}

// HandleLogin handles POST /api/v1/auth/login
func (h *AuthHandler) HandleLogin(ctx context.Context, input *dto.LoginInput) (*dto.LoginOutput, error) {
	result, err := h.authService.Login(ctx, service.LoginInput{
		Email:    input.Body.Email,
		Password: input.Body.Password,
		// TODO: Extract UserAgent and IPAddress from request headers if needed
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.LoginOutput{
		SetCookie: h.createRefreshTokenCookie(result.RefreshToken),
		Body: dto.LoginResponseBody{
			AccessToken: result.AccessToken,
			ExpiresAt:   result.ExpiresAt,
			User:        toUserDTO(result.User),
			Account:     toAccountDTO(result.Account),
		},
	}, nil
}

func (h *AuthHandler) HandleAccountPasswordUpdate(ctx context.Context, input *dto.UpdateAccountPasswordInput) (*dto.UpdateAccountPasswordOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if accountID == contextkeys.NilUUID {
		return nil, apperrors.UnauthorizedError("iam.errors.unauthorized")
	}

	err := h.authService.UpdateAccountPassword(ctx, accountID, service.UpdateAccountPasswordInput{
		ExistingPassword: input.Body.ExistingPassword,
		NewPassword:      input.Body.NewPassword,
		ConfirmPassword:  input.Body.ConfirmPassword,
	})

	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.UpdateAccountPasswordOutput{
		Body: dto.UpdateAccountPasswordResponseBody{
			Message: i18n.Resolve("iam.successes.passwordUpdated", i18n.LocaleFromContext(ctx)),
		},
	}, nil

}

func (h *AuthHandler) HandleGetCurrentUser(ctx context.Context, input *dto.GetCurrentUserInput) (*dto.GetCurrentUserOutput, error) {

	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if userID == contextkeys.NilUUID {
		return nil, apperrors.UnauthorizedError("iam.errors.unauthorized")
	}

	result, err := h.authService.GetCurrentUser(ctx, userID, accountID)

	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.GetCurrentUserOutput{
		Body: dto.GetCurrentUserResponseBody{
			User:    toUserDTO(result.User),
			Account: toAccountDTO(result.Account),
		},
	}, nil
}

// HandleRefresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) HandleRefresh(ctx context.Context, input *dto.RefreshInput) (*dto.RefreshOutput, error) {
	if input.RefreshToken == "" {
		return nil, apperrors.ToHumaError(ctx, apperrors.UnauthorizedError("iam.errors.missingRefreshToken"))
	}

	result, err := h.authService.Refresh(ctx, input.RefreshToken)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.RefreshOutput{
		SetCookie: h.createRefreshTokenCookie(result.RefreshToken),
		Body: dto.RefreshResponseBody{
			AccessToken: result.AccessToken,
			ExpiresAt:   result.ExpiresAt,
		},
	}, nil
}

// HandleLogout handles POST /api/v1/auth/logout
// Requires authentication (SessionID from middleware context).
func (h *AuthHandler) HandleLogout(ctx context.Context, _ *dto.LogoutInput) (*dto.LogoutOutput, error) {
	sessionID := contextkeys.GetSessionID(ctx.Value(contextkeys.SessionID))
	if sessionID == contextkeys.NilUUID {
		return nil, apperrors.ToHumaError(ctx, apperrors.UnauthorizedError("iam.errors.unauthorized"))
	}

	if err := h.authService.Logout(ctx, sessionID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.LogoutOutput{
		SetCookie: h.createClearRefreshTokenCookie(),
	}, nil
}

// HandleLogoutAll handles POST /api/v1/auth/logout/all
// Requires authentication (SessionID from middleware context).
func (h *AuthHandler) HandleLogoutAll(ctx context.Context, _ *dto.LogoutAllInput) (*dto.LogoutAllOutput, error) {
	sessionID := contextkeys.GetSessionID(ctx.Value(contextkeys.SessionID))
	if sessionID == contextkeys.NilUUID {
		return nil, apperrors.ToHumaError(ctx, apperrors.UnauthorizedError("iam.errors.unauthorized"))
	}

	// Get AccountID from SessionID
	accountID, err := h.authService.GetAccountIDBySessionID(ctx, sessionID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	if err := h.authService.LogoutAll(ctx, accountID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.LogoutAllOutput{
		SetCookie: h.createClearRefreshTokenCookie(),
	}, nil
}

// createRefreshTokenCookie creates a secure HttpOnly cookie for the refresh token.
func (h *AuthHandler) createRefreshTokenCookie(refreshToken string) http.Cookie {
	return http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     refreshCookiePath,
		HttpOnly: true,
		Secure:   h.isSecureCookie(),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(h.tokenService.GetRefreshTokenTTL().Seconds()),
	}
}

// createClearRefreshTokenCookie creates a cookie that clears the refresh token.
func (h *AuthHandler) createClearRefreshTokenCookie() http.Cookie {
	return http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     refreshCookiePath,
		HttpOnly: true,
		Secure:   h.isSecureCookie(),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1, // Instructs browser to delete the cookie
		Expires:  time.Unix(0, 0),
	}
}

func (h *AuthHandler) isSecureCookie() bool {
	if h.cfg == nil {
		return false
	}
	return h.cfg.IsProduction()
}

// toUserDTO maps a domain User entity to a UserDTO.
func toUserDTO(user *entity.User) dto.UserDTO {
	return dto.UserDTO{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		ImageURL:  user.ImageURL,
		Bio:       user.Bio,
	}
}

// toAccountDTO maps a domain Account entity to an AccountDTO.
func toAccountDTO(account *entity.Account) dto.AccountDTO {
	return dto.AccountDTO{
		ID:     account.ID,
		Email:  account.Email,
		Status: string(account.Status),
	}
}
