package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type ThreadBlockedUserRepository interface {
	IsBlocked(ctx context.Context, threadID, accountID uuid.UUID) (bool, error)
	Block(ctx context.Context, threadID, blockedID, blockedByID uuid.UUID, reason *string) error
	Unblock(ctx context.Context, threadID, blockedID uuid.UUID) error
	ListBlocked(ctx context.Context, threadID uuid.UUID, q query.QueryOptions) ([]*entity.ThreadBlockedUser, error)
}
