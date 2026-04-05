package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"sync"
	"time"
)

const (
	stateCookieName = "oauth_state"
	stateTTL        = 10 * time.Minute
	stateLength     = 32
)

var (
	ErrInvalidState  = errors.New("oauth: invalid state")
	ErrExpiredState  = errors.New("oauth: state expired")
	ErrStateNotFound = errors.New("oauth: state not found")
)

type StateData struct {
	Provider    string    `json:"p"`
	IsLinking   bool      `json:"l"`
	AccountID   string    `json:"a,omitempty"`
	SessionData string    `json:"s,omitempty"`
	CreatedAt   time.Time `json:"c"`
}

type CookieConfig struct {
	Name     string
	MaxAge   int
	Path     string
	Domain   string
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite
}

type StateManager struct {
	mu     sync.RWMutex
	states map[string]*StateData
	maxAge time.Duration
	cookie CookieConfig
}

func NewStateManager(cookieDomain string, isProduction bool) *StateManager {
	sm := &StateManager{
		states: make(map[string]*StateData),
		maxAge: stateTTL,
		cookie: CookieConfig{
			Name:     stateCookieName,
			MaxAge:   int(stateTTL.Seconds()),
			Path:     "/",
			Domain:   cookieDomain,
			Secure:   isProduction,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		},
	}

	go sm.cleanupLoop()

	return sm
}

func (sm *StateManager) GenerateState(ctx context.Context, provider string, isLinking bool, accountID string) (string, *http.Cookie, error) {
	b := make([]byte, stateLength)
	if _, err := rand.Read(b); err != nil {
		return "", nil, err
	}
	state := hex.EncodeToString(b)

	data := &StateData{
		Provider:  provider,
		IsLinking: isLinking,
		AccountID: accountID,
		CreatedAt: time.Now(),
	}

	sm.mu.Lock()
	sm.states[state] = data
	sm.mu.Unlock()

	cookie := &http.Cookie{
		Name:     sm.cookie.Name,
		Value:    state,
		MaxAge:   sm.cookie.MaxAge,
		Path:     sm.cookie.Path,
		Domain:   sm.cookie.Domain,
		Secure:   sm.cookie.Secure,
		HttpOnly: sm.cookie.HttpOnly,
		SameSite: sm.cookie.SameSite,
	}

	return state, cookie, nil
}

func (sm *StateManager) ValidateState(ctx context.Context, state string) (*StateData, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	data, exists := sm.states[state]
	if !exists {
		return nil, ErrStateNotFound
	}

	if time.Since(data.CreatedAt) > sm.maxAge {
		delete(sm.states, state)
		return nil, ErrExpiredState
	}

	return data, nil
}

func (sm *StateManager) StoreSessionData(state string, sessionData string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if data, exists := sm.states[state]; exists {
		data.SessionData = sessionData
	}
}

func (sm *StateManager) GetSessionData(state string) string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if data, exists := sm.states[state]; exists {
		return data.SessionData
	}
	return ""
}

func (sm *StateManager) DeleteState(state string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.states, state)
}

func (sm *StateManager) GetStateCookieName() string {
	return sm.cookie.Name
}

func (sm *StateManager) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		sm.cleanup()
	}
}

func (sm *StateManager) cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	cutoff := time.Now().Add(-sm.maxAge)
	for state, data := range sm.states {
		if data.CreatedAt.Before(cutoff) {
			delete(sm.states, state)
		}
	}
}
