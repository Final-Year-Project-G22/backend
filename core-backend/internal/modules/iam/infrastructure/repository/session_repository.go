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

type sessionRepository struct {
	sharedrepo.GenericRepository[entity.Session]
	db     *core.Database
	logger core.Logger
}

// NewSessionRepository creates a new SessionRepository implementation.
func NewSessionRepository(db *core.Database, logger core.Logger) repository.SessionRepository {
	base := sharedrepo.NewBaseRepository[entity.Session](db, logger)
	return &sessionRepository{
		GenericRepository: base,
		db:                db,
		logger:            logger,
	}
}

func (r *sessionRepository) GetByRefreshTokenHash(ctx context.Context, hash string) (*entity.Session, error) {
	var session entity.Session

	err := r.db.WithContext(ctx).
		Where("refresh_token_hash = ?", hash).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", time.Now()).
		First(&session).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, iamerror.ErrSessionNotFound
		}
		r.logger.Error("Failed to get session by refresh token hash", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	return &session, nil
}

func (r *sessionRepository) ListActiveByAccountID(ctx context.Context, accountID uuid.UUID) ([]*entity.Session, error) {
	var sessions []*entity.Session

	err := r.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", time.Now()).
		Find(&sessions).Error

	if err != nil {
		r.logger.Error("Failed to list active sessions by account ID", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	return sessions, nil
}

func (r *sessionRepository) RevokeByID(ctx context.Context, id uuid.UUID, revokedAt time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&entity.Session{}).
		Where("id = ?", id).
		Where("revoked_at IS NULL").
		Update("revoked_at", revokedAt)

	if result.Error != nil {
		r.logger.Error("Failed to revoke session", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}

	if result.RowsAffected == 0 {
		return iamerror.ErrSessionNotFound
	}

	return nil
}

func (r *sessionRepository) RevokeAllByAccountID(ctx context.Context, accountID uuid.UUID, revokedAt time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&entity.Session{}).
		Where("account_id = ?", accountID).
		Where("revoked_at IS NULL").
		Update("revoked_at", revokedAt)

	if result.Error != nil {
		r.logger.Error("Failed to revoke all sessions by account ID", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}

	return nil
}

func (r *sessionRepository) DeleteExpired(ctx context.Context, now time.Time) error {
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", now).
		Delete(&entity.Session{})

	if result.Error != nil {
		r.logger.Error("Failed to delete expired sessions", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}

	r.logger.Info("Deleted expired sessions", core.Int64("count", result.RowsAffected))
	return nil
}
