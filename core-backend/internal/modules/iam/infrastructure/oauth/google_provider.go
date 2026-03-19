package oauth

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/markbates/goth/providers/google"
)

type GoogleProvider struct {
	*google.Provider
	encryptor *TokenEncryptor
	logger    core.Logger
}

func NewGoogleProvider(cfg *core.OAuthProviderConfig, encryptor *TokenEncryptor, logger core.Logger) (*GoogleProvider, error) {
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

var _ OAuthProvider = (*GoogleProvider)(nil)
