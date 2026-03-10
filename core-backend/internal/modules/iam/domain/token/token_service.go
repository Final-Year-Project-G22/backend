package token

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AccessTokenClaims struct {
	SessionID uuid.UUID
	Email     string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type TokenService interface {
	// GenerateAccessToken creates a new access token for the given claims.
	GenerateAccessToken(ctx context.Context, claims AccessTokenClaims) (string, error)

	// GenerateRefreshToken creates a cryptographically secure refresh token.
	// Returns the raw token (to send to client) and its SHA-256 hash (to store in DB).
	GenerateRefreshToken(ctx context.Context) (rawToken string, tokenHash string, err error)

	// ValidateAccessToken parses and validates an access token, returning the claims.
	// Returns an error if the token is invalid, expired, or malformed.
	ValidateAccessToken(ctx context.Context, tokenString string) (*AccessTokenClaims, error)

	// HashRefreshToken computes the SHA-256 hash of a refresh token.
	// Used to look up sessions by the token presented by the client.
	HashRefreshToken(rawToken string) string

	// GetAccessTokenTTL returns the configured access token time-to-live.
	GetAccessTokenTTL() time.Duration

	// GetRefreshTokenTTL returns the configured refresh token time-to-live.
	GetRefreshTokenTTL() time.Duration
}
