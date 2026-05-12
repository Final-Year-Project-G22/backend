package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type UserThreadSettingsRepository interface {
	Get(ctx context.Context, accountID, threadID uuid.UUID) (*entity.UserThreadSettings, error)
	ListFollowed(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.UserThreadSettings, error)
	ListFollowStatus(ctx context.Context, accountID uuid.UUID, threadIDs []uuid.UUID) (map[uuid.UUID]bool, error)
	UpsertFollow(ctx context.Context, accountID, threadID uuid.UUID, following bool) error
	SetMuted(ctx context.Context, accountID, threadID uuid.UUID, muted bool) error
	UpdateLastRead(ctx context.Context, accountID, threadID uuid.UUID, at time.Time) error
	Delete(ctx context.Context, accountID, threadID uuid.UUID) error
}
