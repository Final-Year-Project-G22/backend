package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/dto"
	iamdomain "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/oauth"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/token"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
)

const oauthStateCookiePath = "/"

type OAuthHandler struct {
	oauthService iamdomain.OAuthService
	tokenService token.TokenService
	cfg          *core.Config
}

func NewOAuthHandler(oauthService iamdomain.OAuthService, tokenService token.TokenService, cfg *core.Config) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
		tokenService: tokenService,
		cfg:          cfg,
	}
}

// HandleGetProviders handles GET /api/v1/auth/oauth/providers
func (h *OAuthHandler) HandleGetProviders(ctx context.Context, _ *struct{}) (*struct {
	Body struct {
		Providers []dto.OAuthProviderDTO `json:"providers"`
	}
}, error) {
	providers := h.oauthService.GetProviders(ctx)
	result := make([]dto.OAuthProviderDTO, len(providers))
	for i, p := range providers {
		result[i] = dto.OAuthProviderDTO{
			Name:        p.Name,
			DisplayName: p.DisplayName,
			Icon:        p.Icon,
		}
	}
	return &struct {
		Body struct {
			Providers []dto.OAuthProviderDTO `json:"providers"`
		}
	}{
		Body: struct {
			Providers []dto.OAuthProviderDTO `json:"providers"`
		}{
			Providers: result,
		},
	}, nil
}

// HandleInitiateLogin handles GET /api/v1/auth/oauth/login/{provider}
// Redirects to the OAuth provider for authentication.
func (h *OAuthHandler) HandleInitiateLogin(ctx context.Context, input *struct {
	Provider string `path:"provider"`
}) (*struct {
	Status    int
	Location  string       `header:"Location"`
	SetCookie *http.Cookie `header:"Set-Cookie"`
}, error) {
	authURL, cookie, err := h.oauthService.InitiateLogin(ctx, iamdomain.ProviderType(input.Provider))
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	if cookie != nil {
		cookie.Path = oauthStateCookiePath
		cookie.HttpOnly = true
		cookie.Secure = h.isSecureCookie()
		cookie.SameSite = http.SameSiteLaxMode
	}

	return &struct {
		Status    int
		Location  string       `header:"Location"`
		SetCookie *http.Cookie `header:"Set-Cookie"`
	}{
		Status:    http.StatusFound,
		Location:  authURL,
		SetCookie: cookie,
	}, nil
}

// HandleCallback handles GET /api/v1/auth/oauth/callback/{provider}
func (h *OAuthHandler) HandleCallback(ctx context.Context, input *struct {
	Provider string `path:"provider"`
	Code     string `query:"code"`
	State    string `query:"state"`
}) (*dto.OAuthCallbackOutput, error) {
	callbackResult, emailRequired, err := h.oauthService.HandleCallback(ctx, iamdomain.ProviderType(input.Provider), input.Code, input.State)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	if emailRequired != nil {
		return &dto.OAuthCallbackOutput{
			SetCookie: http.Cookie{Name: "oauth_state", Value: "", MaxAge: -1},
			Body: dto.OAuthCallbackResponse{
				User: &dto.UserDTO{
					FirstName: emailRequired.PartialUserInfo.FirstName,
					LastName:  emailRequired.PartialUserInfo.LastName,
				},
				EmailRequired: &dto.OAuthEmailRequiredResponse{
					Provider:   string(emailRequired.PartialUserInfo.Provider),
					Subject:    emailRequired.PartialUserInfo.Subject,
					Name:       emailRequired.PartialUserInfo.Name,
					FirstName:  emailRequired.PartialUserInfo.FirstName,
					LastName:   emailRequired.PartialUserInfo.LastName,
					PictureURL: emailRequired.PartialUserInfo.PictureURL,
					State:      emailRequired.State,
				},
			},
		}, nil
	}

	return &dto.OAuthCallbackOutput{
		SetCookie: h.createRefreshTokenCookie(callbackResult.RefreshToken),
		Body: dto.OAuthCallbackResponse{
			AccessToken:  callbackResult.AccessToken,
			RefreshToken: callbackResult.RefreshToken,
			ExpiresAt:    callbackResult.ExpiresAt,
			User: &dto.UserDTO{
				ID:        callbackResult.User.ID,
				FirstName: callbackResult.User.FirstName,
				LastName:  callbackResult.User.LastName,
				ImageURL:  callbackResult.User.ImageURL,
				Bio:       callbackResult.User.Bio,
			},
			Account: &dto.AccountDTO{
				ID:     callbackResult.Account.ID,
				Email:  callbackResult.Account.Email,
				Status: string(callbackResult.Account.Status),
			},
			IsNewUser: callbackResult.IsNewUser,
		},
	}, nil
}

// HandleCallbackRedirect handles GET /api/v1/auth/oauth/callback/{provider}/mobile
// Returns HTTP 302 redirect to mobile deep-link with tokens in URL.
func (h *OAuthHandler) HandleCallbackRedirect(ctx context.Context, input *struct {
	Provider string `path:"provider"`
	Code     string `query:"code"`
	State    string `query:"state"`
}) (*dto.OAuthCallbackRedirectOutput, error) {
	callbackResult, emailRequired, err := h.oauthService.HandleCallback(ctx, iamdomain.ProviderType(input.Provider), input.Code, input.State)
	if err != nil {
		return h.redirectError(err)
	}

	if emailRequired != nil {
		return h.redirectEmailRequired(emailRequired)
	}

	return h.redirectSuccess(callbackResult)
}

// HandleCompleteWithEmail handles POST /api/v1/auth/oauth/complete
func (h *OAuthHandler) HandleCompleteWithEmail(ctx context.Context, input *dto.OAuthCompleteEmailInput) (*dto.OAuthCallbackOutput, error) {
	result, err := h.oauthService.CompleteWithEmail(ctx, input.Body.State, input.Body.Email)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.OAuthCallbackOutput{
		SetCookie: h.createRefreshTokenCookie(result.RefreshToken),
		Body: dto.OAuthCallbackResponse{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
			ExpiresAt:    result.ExpiresAt,
			User: &dto.UserDTO{
				ID:        result.User.ID,
				FirstName: result.User.FirstName,
				LastName:  result.User.LastName,
				ImageURL:  result.User.ImageURL,
				Bio:       result.User.Bio,
			},
			Account: &dto.AccountDTO{
				ID:     result.Account.ID,
				Email:  result.Account.Email,
				Status: string(result.Account.Status),
			},
			IsNewUser: result.IsNewUser,
		},
	}, nil
}

// HandleInitiateLink handles GET /api/v1/auth/oauth/link/{provider}
// Requires authentication - links OAuth provider to existing account.
func (h *OAuthHandler) HandleInitiateLink(ctx context.Context, input *struct {
	Provider string `path:"provider"`
}) (*struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
}, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if accountID == contextkeys.NilUUID {
		return nil, apperrors.ToHumaError(ctx, apperrors.UnauthorizedError("iam.errors.unauthorized"))
	}

	_, cookie, err := h.oauthService.InitiateLink(ctx, iamdomain.ProviderType(input.Provider), accountID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	if cookie != nil {
		cookie.Path = oauthStateCookiePath
		cookie.HttpOnly = true
		cookie.Secure = h.isSecureCookie()
		cookie.SameSite = http.SameSiteLaxMode
	}

	return &struct {
		SetCookie http.Cookie `header:"Set-Cookie"`
	}{
		SetCookie: *cookie,
	}, nil
}

// HandleLinkCallback handles GET /api/v1/auth/oauth/link/callback/{provider}
func (h *OAuthHandler) HandleLinkCallback(ctx context.Context, input *struct {
	Provider string `path:"provider"`
	Code     string `query:"code"`
	State    string `query:"state"`
}) (*dto.OAuthLinkCallbackOutput, error) {
	linkResult, emailRequiredResult, err := h.oauthService.HandleLinkCallback(ctx, iamdomain.ProviderType(input.Provider), input.Code, input.State)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	if emailRequiredResult != nil {
		return &dto.OAuthLinkCallbackOutput{
			Body: struct {
				Provider string `json:"provider"`
			}{
				Provider: string(emailRequiredResult.PartialUserInfo.Provider),
			},
			EmailRequired: &dto.OAuthEmailRequiredResponse{
				Provider:   string(emailRequiredResult.PartialUserInfo.Provider),
				Subject:    emailRequiredResult.PartialUserInfo.Subject,
				Name:       emailRequiredResult.PartialUserInfo.Name,
				FirstName:  emailRequiredResult.PartialUserInfo.FirstName,
				LastName:   emailRequiredResult.PartialUserInfo.LastName,
				PictureURL: emailRequiredResult.PartialUserInfo.PictureURL,
				State:      emailRequiredResult.State,
			},
		}, nil
	}

	return &dto.OAuthLinkCallbackOutput{
		Body: struct {
			Provider string `json:"provider"`
		}{
			Provider: linkResult.OAuthIdentity.Provider,
		},
	}, nil
}

// HandleGetIdentities handles GET /api/v1/auth/oauth/identities
func (h *OAuthHandler) HandleGetIdentities(ctx context.Context, _ *struct{}) (*dto.OAuthIdentitiesOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if accountID == contextkeys.NilUUID {
		return nil, apperrors.ToHumaError(ctx, apperrors.UnauthorizedError("iam.errors.unauthorized"))
	}

	identities, err := h.oauthService.GetLinkedIdentities(ctx, accountID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	result := make([]dto.OAuthIdentityDTO, len(identities))
	for i, ident := range identities {
		result[i] = dto.OAuthIdentityDTO{
			Provider:   ident.Provider,
			LinkedAt:   ident.CreatedAt,
			LastUsedAt: ident.LastUsedAt,
		}
		if ident.ProviderEmail != nil {
			result[i].ProviderEmail = *ident.ProviderEmail
		}
	}

	return &dto.OAuthIdentitiesOutput{
		Body: dto.OAuthIdentitiesResponse{
			Identities: result,
		},
	}, nil
}

// HandleUnlink handles DELETE /api/v1/auth/oauth/identities/{provider}
func (h *OAuthHandler) HandleUnlink(ctx context.Context, input *struct {
	Provider string `path:"provider"`
}) (*dto.OAuthUnlinkOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if accountID == contextkeys.NilUUID {
		return nil, apperrors.ToHumaError(ctx, apperrors.UnauthorizedError("iam.errors.unauthorized"))
	}

	if err := h.oauthService.UnlinkProvider(ctx, accountID, input.Provider); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.OAuthUnlinkOutput{
		Body: struct {
			Unlinked string `json:"unlinked"`
		}{
			Unlinked: input.Provider,
		},
	}, nil
}

func (h *OAuthHandler) createRefreshTokenCookie(refreshToken string) http.Cookie {
	return http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/api/v1/auth/refresh",
		HttpOnly: true,
		Secure:   h.isSecureCookie(),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(h.tokenService.GetRefreshTokenTTL().Seconds()),
	}
}

func (h *OAuthHandler) isSecureCookie() bool {
	if h.cfg == nil {
		return false
	}
	return h.cfg.IsProduction()
}

func (h *OAuthHandler) redirectError(err error) (*dto.OAuthCallbackRedirectOutput, error) {
	base := h.cfg.OAuth.MobileRedirectBaseURL
	redirectURL := fmt.Sprintf("%s?error=%s", base, url.QueryEscape("oauth_error"))
	return &dto.OAuthCallbackRedirectOutput{
		Status:    http.StatusFound,
		Location:  redirectURL,
		SetCookie: http.Cookie{Name: "oauth_state", Value: "", MaxAge: -1},
	}, nil
}

func (h *OAuthHandler) redirectEmailRequired(result *iamdomain.EmailRequiredResult) (*dto.OAuthCallbackRedirectOutput, error) {
	base := h.cfg.OAuth.MobileRedirectBaseURL
	redirectURL := fmt.Sprintf("%s?email_required=true&provider=%s&state=%s",
		base,
		url.QueryEscape(string(result.PartialUserInfo.Provider)),
		url.QueryEscape(result.State),
	)
	return &dto.OAuthCallbackRedirectOutput{
		Status:    http.StatusFound,
		Location:  redirectURL,
		SetCookie: http.Cookie{Name: "oauth_state", Value: "", MaxAge: -1},
	}, nil
}

func (h *OAuthHandler) redirectSuccess(result *iamdomain.OAuthCallbackResult) (*dto.OAuthCallbackRedirectOutput, error) {
	base := h.cfg.OAuth.MobileRedirectBaseURL
	redirectURL := fmt.Sprintf("%s?access_token=%s&refresh_token=%s&expires_at=%s&is_new_user=%t",
		base,
		url.QueryEscape(result.AccessToken),
		url.QueryEscape(result.RefreshToken),
		url.QueryEscape(result.ExpiresAt.Format(time.RFC3339)),
		result.IsNewUser,
	)
	return &dto.OAuthCallbackRedirectOutput{
		Status:    http.StatusFound,
		Location:  redirectURL,
		SetCookie: http.Cookie{Name: "oauth_state", Value: "", MaxAge: -1},
	}, nil
}
