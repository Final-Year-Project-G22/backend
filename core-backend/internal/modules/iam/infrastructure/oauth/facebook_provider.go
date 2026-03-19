package oauth

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/facebook"
)

type FacebookProvider struct {
	provider  *facebook.Provider
	encryptor *TokenEncryptor
	logger    core.Logger
}

func NewFacebookProvider(cfg *core.OAuthProviderConfig, encryptor *TokenEncryptor, logger core.Logger) (*FacebookProvider, error) {
	provider := facebook.New(cfg.ClientID, cfg.ClientSecret, cfg.RedirectURI, cfg.Scopes...)

	return &FacebookProvider{
		provider:  provider,
		encryptor: encryptor,
		logger:    logger,
	}, nil
}

func (p *FacebookProvider) Name() string {
	return "facebook"
}

func (p *FacebookProvider) GetAuthURL(state string) string {
	session, _ := p.provider.BeginAuth(state)
	if session == nil {
		return ""
	}
	authURL, _ := session.GetAuthURL()
	return authURL
}

func (p *FacebookProvider) BeginAuth(state string) (goth.Session, error) {
	return p.provider.BeginAuth(state)
}

func (p *FacebookProvider) UnmarshalSession(data string) (goth.Session, error) {
	return p.provider.UnmarshalSession(data)
}
