package notification

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type AccountReaderAdapter struct {
	accountRepo repository.AccountRepository
}

func NewAccountReaderAdapter(accountRepo repository.AccountRepository) *AccountReaderAdapter {
	return &AccountReaderAdapter{accountRepo: accountRepo}
}

func (a *AccountReaderAdapter) FindAll(ctx context.Context) ([]uuid.UUID, error) {
	accounts, err := a.accountRepo.Find(ctx, query.DefaultQueryOptions())
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, len(accounts))
	for i, acct := range accounts {
		ids[i] = acct.ID
	}
	return ids, nil
}

func (a *AccountReaderAdapter) FindBySegment(ctx context.Context, segment map[string]interface{}) ([]uuid.UUID, error) {
	accounts, err := a.accountRepo.FindBySegment(ctx, segment)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, len(accounts))
	for i, acct := range accounts {
		ids[i] = acct.ID
	}
	return ids, nil
}
