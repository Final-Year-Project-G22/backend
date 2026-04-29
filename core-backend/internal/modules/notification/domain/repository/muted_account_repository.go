package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type MutedAccountRepository interface {
	sharedrepo.GenericRepository[entity.MutedAccount]

	IsMuted(ctx context.Context, accountID uuid.UUID, mutedAccountID uuid.UUID) (bool, error)
	ListByAccount(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.MutedAccount, error)
	DeleteByAccountPair(ctx context.Context, accountID uuid.UUID, mutedAccountID uuid.UUID) error
}
