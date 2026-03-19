package oauth

import (
	"github.com/markbates/goth"
)

type OAuthProvider interface {
	Name() string
	GetAuthURL(state string) string
	BeginAuth(state string) (goth.Session, error)
	UnmarshalSession(data string) (goth.Session, error)
}

type ProviderRegistry struct {
	providers map[string]OAuthProvider
}

func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]OAuthProvider),
	}
}

func (r *ProviderRegistry) Register(provider OAuthProvider) {
	r.providers[provider.Name()] = provider
}

func (r *ProviderRegistry) Get(name string) (OAuthProvider, bool) {
	p, ok := r.providers[name]
	return p, ok
}

func (r *ProviderRegistry) List() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}
