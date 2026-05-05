package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	iamerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type accountEmailOTPRepository struct {
	sharedrepo.GenericRepository[entity.AccountEmailOTP]
	db     *core.Database
	logger core.Logger
}

func NewAccountEmailOTPRepository(db *core.Database, logger core.Logger) repository.AccountEmailOTPRepository {
	base := sharedrepo.NewBaseRepository[entity.AccountEmailOTP](db, logger)
	return &accountEmailOTPRepository{
		GenericRepository: base,
		db:                db,
		logger:            logger,
	}
}

func (r *accountEmailOTPRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *accountEmailOTPRepository) GetLatestByAccountID(ctx context.Context, accountID uuid.UUID) (*entity.AccountEmailOTP, error) {
	var otp entity.AccountEmailOTP
	err := r.getDB(ctx).
		Where("account_id = ?", accountID).
		Order("created_at DESC").
		First(&otp).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, iamerror.ErrEmailOTPNotFound
		}
		r.logger.Error("Failed to get latest email OTP", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	return &otp, nil
}

func (r *accountEmailOTPRepository) GetActiveByAccountID(ctx context.Context, accountID uuid.UUID, now time.Time) (*entity.AccountEmailOTP, error) {
	var otp entity.AccountEmailOTP
	err := r.getDB(ctx).
		Where("account_id = ?", accountID).
		Where("consumed_at IS NULL").
		Where("expires_at > ?", now).
		Order("created_at DESC").
		First(&otp).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, iamerror.ErrEmailOTPNotFound
		}
		r.logger.Error("Failed to get active email OTP", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	return &otp, nil
}

func (r *accountEmailOTPRepository) GetActiveByAccountIDAndPurpose(ctx context.Context, accountID uuid.UUID, purpose string, now time.Time) (*entity.AccountEmailOTP, error) {
	var otp entity.AccountEmailOTP
	err := r.getDB(ctx).
		Where("account_id = ?", accountID).
		Where("purpose = ?", purpose).
		Where("consumed_at IS NULL").
		Where("expires_at > ?", now).
		Order("created_at DESC").
		First(&otp).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, iamerror.ErrEmailOTPNotFound
		}
		r.logger.Error("Failed to get active email OTP by purpose", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	return &otp, nil
}

func (r *accountEmailOTPRepository) IncrementAttemptCount(ctx context.Context, otpID uuid.UUID) error {
	result := r.getDB(ctx).
		Model(&entity.AccountEmailOTP{}).
		Where("id = ?", otpID).
		Update("attempt_count", gorm.Expr("attempt_count + 1"))
	if result.Error != nil {
		r.logger.Error("Failed to increment email OTP attempt count", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return iamerror.ErrEmailOTPNotFound
	}

	return nil
}

func (r *accountEmailOTPRepository) Consume(ctx context.Context, otpID uuid.UUID, consumedAt time.Time) error {
	result := r.getDB(ctx).
		Model(&entity.AccountEmailOTP{}).
		Where("id = ?", otpID).
		Where("consumed_at IS NULL").
		Update("consumed_at", consumedAt)
	if result.Error != nil {
		r.logger.Error("Failed to consume email OTP", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return iamerror.ErrEmailOTPAlreadyConsumed
	}

	return nil
}

func (r *accountEmailOTPRepository) InvalidateActiveByAccountID(ctx context.Context, accountID uuid.UUID, now time.Time) error {
	result := r.getDB(ctx).
		Model(&entity.AccountEmailOTP{}).
		Where("account_id = ?", accountID).
		Where("consumed_at IS NULL").
		Where("expires_at > ?", now).
		Update("consumed_at", now)
	if result.Error != nil {
		r.logger.Error("Failed to invalidate active email OTP", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}

	return nil
}

func (r *accountEmailOTPRepository) InvalidateActiveByAccountIDAndPurpose(ctx context.Context, accountID uuid.UUID, purpose string, now time.Time) error {
	result := r.getDB(ctx).
		Model(&entity.AccountEmailOTP{}).
		Where("account_id = ?", accountID).
		Where("purpose = ?", purpose).
		Where("consumed_at IS NULL").
		Where("expires_at > ?", now).
		Update("consumed_at", now)
	if result.Error != nil {
		r.logger.Error("Failed to invalidate active email OTP by purpose", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}

	return nil
}

func (r *accountEmailOTPRepository) IncrementResendCountAndUpdateLastSentAt(ctx context.Context, otpID uuid.UUID, now time.Time) error {
	result := r.getDB(ctx).
		Model(&entity.AccountEmailOTP{}).
		Where("id = ?", otpID).
		Updates(map[string]interface{}{
			"resend_count": gorm.Expr("resend_count + 1"),
			"last_sent_at": now,
		})
	if result.Error != nil {
		r.logger.Error("Failed to update resend metadata for email OTP", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return iamerror.ErrEmailOTPNotFound
	}

	return nil
}

func (r *accountEmailOTPRepository) FindActiveByCodeHashAndPurpose(ctx context.Context, codeHash string, purpose string, now time.Time) (*entity.AccountEmailOTP, error) {
	var otp entity.AccountEmailOTP
	err := r.getDB(ctx).
		Where("code_hash = ?", codeHash).
		Where("purpose = ?", purpose).
		Where("consumed_at IS NULL").
		Where("expires_at > ?", now).
		Order("created_at DESC").
		First(&otp).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, iamerror.ErrEmailOTPNotFound
		}
		r.logger.Error("Failed to find active OTP by code hash and purpose", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &otp, nil
}
