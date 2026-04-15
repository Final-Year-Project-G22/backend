package usecase

import (
	"context"

	"github.com/google/uuid"
)

type ThreadBlockUsecase interface {
	BlockUser(ctx context.Context, input BlockUserInput) error
	UnblockUser(ctx context.Context, input BlockUserInput) error
	IsBlocked(ctx context.Context, threadID, accountID uuid.UUID) (bool, error)
}
