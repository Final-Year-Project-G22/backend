package oauth

import (
	"log"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/facebook"
)

type FacebookProvider struct {
	*facebook.Provider
	encryptor *TokenEncryptor
	logger    core.Logger
}

func NewFacebookProvider(cfg *core.OAuthProviderConfig, encryptor *TokenEncryptor, logger core.Logger) (*FacebookProvider, error) {
	if cfg == nil || cfg.ClientID == "" || cfg.ClientSecret == "" {
		log.Println("OAuth: Facebook provider not configured (missing client_id or client_secret)")
		return nil, nil
	}

	provider := facebook.New(cfg.ClientID, cfg.ClientSecret, cfg.RedirectURI, cfg.Scopes...)

	return &FacebookProvider{
		Provider:  provider,
		encryptor: encryptor,
		logger:    logger,
	}, nil
}

func (p *FacebookProvider) Name() string {
	return "facebook"
}

func (p *FacebookProvider) GetAuthURL(state string) string {
	session, _ := p.BeginAuth(state)
	if session == nil {
		return ""
	}
	authURL, _ := session.GetAuthURL()
	return authURL
}

func (p *FacebookProvider) GetGothProvider() goth.Provider {
	return p.Provider
}

var _ OAuthProvider = (*FacebookProvider)(nil)
