package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type UserCategorySettingsRepository interface {
	Get(ctx context.Context, accountID, categoryID uuid.UUID) (*entity.UserCategorySettings, error)
	ListFollowed(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.UserCategorySettings, error)
	UpsertFollow(ctx context.Context, accountID, categoryID uuid.UUID, following bool) error
	SetMuted(ctx context.Context, accountID, categoryID uuid.UUID, muted bool) error
	UpdateLastRead(ctx context.Context, accountID, categoryID uuid.UUID, at time.Time) error
	Delete(ctx context.Context, accountID, categoryID uuid.UUID) error
}
