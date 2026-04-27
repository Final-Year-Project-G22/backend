package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notiferror "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/error"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type mutedAccountRepository struct {
	sharedrepo.GenericRepository[entity.MutedAccount]
	db     *core.Database
	logger core.Logger
}

func NewMutedAccountRepository(db *core.Database, logger core.Logger) notifrepo.MutedAccountRepository {
	base := sharedrepo.NewBaseRepository[entity.MutedAccount](db, logger)
	return &mutedAccountRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *mutedAccountRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *mutedAccountRepository) IsMuted(ctx context.Context, accountID uuid.UUID, mutedAccountID uuid.UUID) (bool, error) {
	var muted entity.MutedAccount
	err := r.getDB(ctx).
		Where("account_id = ? AND muted_account_id = ? AND (mute_until IS NULL OR mute_until > ?)", accountID, mutedAccountID, time.Now()).
		First(&muted).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		r.logger.Error("Failed to check mute status", core.Error(err))
		return false, errors.InternalError("errors.databaseError", err)
	}
	return true, nil
}

func (r *mutedAccountRepository) ListByAccount(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.MutedAccount, error) {
	var mutes []*entity.MutedAccount
	db := r.getDB(ctx).Where("account_id = ?", accountID)
	db = applyPaginationAndSorting(db, q, "created_at desc")
	if err := db.Find(&mutes).Error; err != nil {
		r.logger.Error("Failed to list muted accounts", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return mutes, nil
}

func (r *mutedAccountRepository) DeleteByAccountPair(ctx context.Context, accountID uuid.UUID, mutedAccountID uuid.UUID) error {
	result := r.getDB(ctx).
		Where("account_id = ? AND muted_account_id = ?", accountID, mutedAccountID).
		Delete(&entity.MutedAccount{})
	if result.Error != nil {
		r.logger.Error("Failed to delete mute", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return notiferror.ErrMuteNotFound
	}
	return nil
}
