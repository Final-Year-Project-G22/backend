package oauth

import (
	"context"
	"net/http"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/google/uuid"
)

type ProviderType string

const (
	ProviderGoogle   ProviderType = "google"
	ProviderFacebook ProviderType = "facebook"
)

type ProviderInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Icon        string `json:"icon"`
}

type OAuthUserInfo struct {
	Provider      ProviderType `json:"-"`
	Subject       string       `json:"subject"`
	Email         string       `json:"email,omitempty"`
	EmailVerified bool         `json:"-"`
	Name          string       `json:"name"`
	FirstName     string       `json:"firstName"`
	LastName      string       `json:"lastName"`
	PictureURL    string       `json:"pictureUrl,omitempty"`
}

type OAuthCallbackResult struct {
	IsNewUser    bool            `json:"isNewUser"`
	AccessToken  string          `json:"accessToken"`
	RefreshToken string          `json:"refreshToken"`
	ExpiresAt    time.Time       `json:"expiresAt"`
	User         *entity.User    `json:"user"`
	Account      *entity.Account `json:"account"`
}

type EmailRequiredResult struct {
	PartialUserInfo OAuthUserInfo `json:"partialUserInfo"`
	State           string        `json:"state"`
}

type LinkResult struct {
	OAuthIdentity *entity.OAuthIdentity `json:"oauthIdentity"`
}

type OAuthService interface {
	GetProviders(ctx context.Context) []ProviderInfo

	InitiateLogin(ctx context.Context, provider ProviderType) (string, *http.Cookie, error)
	InitiateLink(ctx context.Context, provider ProviderType, accountID uuid.UUID) (string, *http.Cookie, error)

	HandleCallback(ctx context.Context, provider ProviderType, code string, state string) (*OAuthCallbackResult, *EmailRequiredResult, error)
	HandleLinkCallback(ctx context.Context, provider ProviderType, code string, state string) (*LinkResult, *EmailRequiredResult, error)

	CompleteWithEmail(ctx context.Context, state string, email string) (*OAuthCallbackResult, error)

	GetLinkedIdentities(ctx context.Context, accountID uuid.UUID) ([]*entity.OAuthIdentity, error)
	UnlinkProvider(ctx context.Context, accountID uuid.UUID, provider string) error
}
