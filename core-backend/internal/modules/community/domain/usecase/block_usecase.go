package usecase

import (
	"context"

	"github.com/google/uuid"
)

type ThreadBlockUsecase interface {
	BlockUser(ctx context.Context, actorID, threadID, blockedID uuid.UUID, reason *string) error
	UnblockUser(ctx context.Context, actorID, threadID, blockedID uuid.UUID) error
	IsBlocked(ctx context.Context, threadID, accountID uuid.UUID) (bool, error)
}
