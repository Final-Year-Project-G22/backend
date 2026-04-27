package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type NotificationMuteUsecase interface {
	MuteAccount(ctx context.Context, accountID uuid.UUID, input MuteAccountInput) error
	UnmuteAccount(ctx context.Context, accountID uuid.UUID, mutedAccountID uuid.UUID) error
	IsMuted(ctx context.Context, accountID uuid.UUID, mutedAccountID uuid.UUID) (bool, error)
	ListMutedAccounts(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.MutedAccount, error)
}
