package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	iamrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type accountPreferenceRepository struct {
	sharedrepo.GenericRepository[entity.AccountPreference]
	db     *core.Database
	logger core.Logger
}

func NewAccountPreferenceRepository(db *core.Database, logger core.Logger) iamrepo.AccountPreferenceRepository {
	base := sharedrepo.NewBaseRepository[entity.AccountPreference](db, logger)
	return &accountPreferenceRepository{
		GenericRepository: base,
		db:                db,
		logger:            logger,
	}
}

func (r *accountPreferenceRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *accountPreferenceRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID) (*entity.AccountPreference, error) {
	var pref entity.AccountPreference
	if err := r.getDB(ctx).
		Where("account_id = ?", accountID).
		First(&pref).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Error("Failed to get account preference by account ID", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &pref, nil
}
