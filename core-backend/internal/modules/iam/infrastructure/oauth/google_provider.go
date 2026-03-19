package oauth

import (
	"log"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
)

type GoogleProvider struct {
	*google.Provider
	encryptor *TokenEncryptor
	logger    core.Logger
}

func NewGoogleProvider(cfg *core.OAuthProviderConfig, encryptor *TokenEncryptor, logger core.Logger) (*GoogleProvider, error) {
	if cfg == nil || cfg.ClientID == "" || cfg.ClientSecret == "" {
		log.Println("OAuth: Google provider not configured (missing client_id or client_secret)")
		return nil, nil
	}

	provider := google.New(cfg.ClientID, cfg.ClientSecret, cfg.RedirectURI, cfg.Scopes...)

	return &GoogleProvider{
		Provider:  provider,
		encryptor: encryptor,
		logger:    logger,
	}, nil
}

func (p *GoogleProvider) Name() string {
	return "google"
}

func (p *GoogleProvider) GetAuthURL(state string) string {
	session, _ := p.BeginAuth(state)
	if session == nil {
		return ""
	}
	authURL, _ := session.GetAuthURL()
	return authURL
}

func (p *GoogleProvider) GetGothProvider() goth.Provider {
	return p.Provider
}

var _ OAuthProvider = (*GoogleProvider)(nil)
