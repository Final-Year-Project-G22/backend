package repository

import (
	"context"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	iamerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type accountRepository struct {
	sharedrepo.GenericRepository[entity.Account]
	db     *core.Database
	logger core.Logger
}

// NewAccountRepository creates a new AccountRepository implementation.
func NewAccountRepository(db *core.Database, logger core.Logger) repository.AccountRepository {
	base := sharedrepo.NewBaseRepository[entity.Account](db, logger)
	return &accountRepository{
		GenericRepository: base,
		db:                db,
		logger:            logger,
	}
}

// getDB returns the appropriate *gorm.DB for the context (tx-aware).
func (r *accountRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *accountRepository) GetByEmailNormalized(ctx context.Context, email string) (*entity.Account, error) {
	var account entity.Account
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	err := r.getDB(ctx).
		Where("email_normalized = ?", normalizedEmail).
		First(&account).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, iamerror.ErrAccountNotFound
		}
		r.logger.Error("Failed to get account by email", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	return &account, nil
}

func (r *accountRepository) GetByUsernameNormalized(ctx context.Context, username string) (*entity.Account, error) {
	var account entity.Account
	normalizedUsername := strings.ToLower(strings.TrimSpace(username))

	err := r.getDB(ctx).
		Where("username_normalized = ?", normalizedUsername).
		First(&account).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, iamerror.ErrAccountNotFound
		}
		r.logger.Error("Failed to get account by username", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	return &account, nil
}

func (r *accountRepository) GetByEmailOrUsername(ctx context.Context, identifier string) (*entity.Account, error) {
	normalized := strings.ToLower(strings.TrimSpace(identifier))

	account, err := r.GetByEmailNormalized(ctx, normalized)
	if err == nil {
		return account, nil
	}
	if err != iamerror.ErrAccountNotFound {
		return nil, err
	}

	return r.GetByUsernameNormalized(ctx, normalized)
}

func (r *accountRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Account, error) {
	var accounts []*entity.Account

	err := r.getDB(ctx).
		Where("user_id = ?", userID).
		Find(&accounts).Error

	if err != nil {
		r.logger.Error("Failed to list accounts by user ID", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	return accounts, nil
}

func (r *accountRepository) ExistsByEmailNormalized(ctx context.Context, email string) (bool, error) {
	var count int64
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	err := r.getDB(ctx).
		Model(&entity.Account{}).
		Where("email_normalized = ?", normalizedEmail).
		Count(&count).Error

	if err != nil {
		r.logger.Error("Failed to check account existence by email", core.Error(err))
		return false, errors.InternalError("errors.databaseError", err)
	}

	return count > 0, nil
}

func (r *accountRepository) ExistsByUsernameNormalized(ctx context.Context, username string) (bool, error) {
	var count int64
	normalizedUsername := strings.ToLower(strings.TrimSpace(username))

	err := r.getDB(ctx).
		Model(&entity.Account{}).
		Where("username_normalized = ?", normalizedUsername).
		Count(&count).Error

	if err != nil {
		r.logger.Error("Failed to check account existence by username", core.Error(err))
		return false, errors.InternalError("errors.databaseError", err)
	}

	return count > 0, nil
}

func (r *accountRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.AccountStatus) error {
	result := r.getDB(ctx).
		Model(&entity.Account{}).
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		r.logger.Error("Failed to update account status", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}

	if result.RowsAffected == 0 {
		return iamerror.ErrAccountNotFound
	}

	return nil
}

func (r *accountRepository) MarkEmailVerifiedAndActivate(ctx context.Context, id uuid.UUID) error {
	result := r.getDB(ctx).
		Model(&entity.Account{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"email_verified": true,
			"status":         entity.AccountStatusActive,
		})

	if result.Error != nil {
		r.logger.Error("Failed to mark account email verified and active", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}

	if result.RowsAffected == 0 {
		return iamerror.ErrAccountNotFound
	}

	return nil
}
