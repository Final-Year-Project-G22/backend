package token

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type jwtClaims struct {
	jwt.RegisteredClaims
	SessionID string `json:"sid"`
	Email     string `json:"email"`
}

type jwtService struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	logger          core.Logger
}

// NewJWTService creates a new TokenService implementation using JWT.
func NewJWTService(cfg *core.Config, logger core.Logger) token.TokenService {
	return &jwtService{
		secret:          []byte(cfg.JWT.Secret),
		accessTokenTTL:  cfg.JWT.AccessTokenTTL,
		refreshTokenTTL: cfg.JWT.RefreshTokenTTL,
		logger:          logger,
	}
}

func (s *jwtService) GenerateAccessToken(ctx context.Context, claims token.AccessTokenClaims) (string, error) {
	now := time.Now()
	expiresAt := now.Add(s.accessTokenTTL)

	jwtClaims := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
		},
		SessionID: claims.SessionID.String(),
		Email:     claims.Email,
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	tokenString, err := t.SignedString(s.secret)
	if err != nil {
		s.logger.Error("Failed to sign access token", core.Error(err))
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return tokenString, nil
}

func (s *jwtService) GenerateRefreshToken(ctx context.Context) (rawToken string, tokenHash string, err error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		s.logger.Error("Failed to generate refresh token", core.Error(err))
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	rawToken = base64.URLEncoding.EncodeToString(tokenBytes)
	tokenHash = s.HashRefreshToken(rawToken)

	return rawToken, tokenHash, nil
}

func (s *jwtService) ValidateAccessToken(ctx context.Context, tokenString string) (*token.AccessTokenClaims, error) {
	t, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := t.Claims.(*jwtClaims)
	if !ok || !t.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	sessionID, err := uuid.Parse(claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session_id in token: %w", err)
	}

	return &token.AccessTokenClaims{
		SessionID: sessionID,
		Email:     claims.Email,
		IssuedAt:  claims.IssuedAt.Time,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

func (s *jwtService) HashRefreshToken(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(hash[:])
}

func (s *jwtService) GetAccessTokenTTL() time.Duration {
	return s.accessTokenTTL
}

func (s *jwtService) GetRefreshTokenTTL() time.Duration {
	return s.refreshTokenTTL
}
