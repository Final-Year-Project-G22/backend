package usecase

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/google/uuid"
)

type SessionUsecase interface {
	CreateSession(ctx context.Context, accountID uuid.UUID, input CreateSessionInput) (*entity.Session, error)
	GetSessionByRefreshTokenHash(ctx context.Context, hash string) (*entity.Session, error)
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
	RevokeAllSessions(ctx context.Context, accountID uuid.UUID) error
	CleanupExpiredSessions(ctx context.Context) error
}

type CreateSessionInput struct {
	RefreshTokenHash string
	UserAgent        *string
	IPAddress        *string
	ExpiresAt        time.Time
}
