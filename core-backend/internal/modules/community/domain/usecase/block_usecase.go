package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type ThreadBlockUsecase interface {
	BlockUser(ctx context.Context, input BlockUserInput) error
	UnblockUser(ctx context.Context, input BlockUserInput) error
	IsBlocked(ctx context.Context, threadID, accountID uuid.UUID) (bool, error)
	ListBlockedUsers(ctx context.Context, threadID uuid.UUID, q query.QueryOptions) ([]*entity.ThreadBlockedUser, int64, error)
	ListAllBlockedUsers(ctx context.Context, q query.QueryOptions) ([]*entity.ThreadBlockedUser, int64, error)
}
