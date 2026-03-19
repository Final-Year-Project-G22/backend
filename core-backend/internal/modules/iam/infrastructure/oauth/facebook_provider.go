package oauth

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/markbates/goth/providers/facebook"
)

type FacebookProvider struct {
	*facebook.Provider
	encryptor *TokenEncryptor
	logger    core.Logger
}

func NewFacebookProvider(cfg *core.OAuthProviderConfig, encryptor *TokenEncryptor, logger core.Logger) (*FacebookProvider, error) {
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

var _ OAuthProvider = (*FacebookProvider)(nil)
