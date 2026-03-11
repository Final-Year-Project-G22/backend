package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type SessionRepository interface {
	sharedrepo.GenericRepository[entity.Session]

	GetActiveByID(ctx context.Context, id uuid.UUID) (*entity.Session, error)
	GetByRefreshTokenHash(ctx context.Context, hash string) (*entity.Session, error)
	ListActiveByAccountID(ctx context.Context, accountID uuid.UUID) ([]*entity.Session, error)
	RevokeByID(ctx context.Context, id uuid.UUID, revokedAt time.Time) error
	RevokeAllByAccountID(ctx context.Context, accountID uuid.UUID, revokedAt time.Time) error
	DeleteExpired(ctx context.Context, now time.Time) error
}
