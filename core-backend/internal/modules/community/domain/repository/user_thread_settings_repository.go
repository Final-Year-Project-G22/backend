package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	"github.com/google/uuid"
)

type UserThreadSettingsRepository interface {
	Get(ctx context.Context, accountID, threadID uuid.UUID) (*entity.UserThreadSettings, error)
	UpsertFollow(ctx context.Context, accountID, threadID uuid.UUID, following bool) error
	SetMuted(ctx context.Context, accountID, threadID uuid.UUID, muted bool) error
	UpdateLastRead(ctx context.Context, accountID, threadID uuid.UUID, at time.Time) error
	Delete(ctx context.Context, accountID, threadID uuid.UUID) error
}
