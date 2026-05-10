package notification

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccountReaderAdapter struct {
	accountRepo repository.AccountRepository
	db          *core.Database
}

func NewAccountReaderAdapter(accountRepo repository.AccountRepository, db *core.Database) *AccountReaderAdapter {
	return &AccountReaderAdapter{accountRepo: accountRepo, db: db}
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

func (a *AccountReaderAdapter) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return a.db.WithContext(ctx)
}

func (a *AccountReaderAdapter) GetAccountInfo(ctx context.Context, accountID uuid.UUID) (*notifrepo.AccountInfo, error) {
	var account entity.Account
	if err := a.getDB(ctx).Preload("AccountPreference").Preload("User").First(&account, "id = ?", accountID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &notifrepo.AccountInfo{Email: "", Locale: "en"}, nil
		}
		return nil, err
	}

	locale := "en"
	if account.AccountPreference != nil {
		locale = account.AccountPreference.Language
	}

	name := ""
	if account.User.FirstName != "" {
		name = account.User.FirstName
	}

	return &notifrepo.AccountInfo{
		Email:  account.Email,
		Locale: locale,
		Name:   name,
	}, nil
}
