package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type notificationOutboxRepository struct {
	db     *core.Database
	logger core.Logger
}

func NewNotificationOutboxRepository(db *core.Database, logger core.Logger) notifrepo.NotificationOutboxRepository {
	return &notificationOutboxRepository{db: db, logger: logger}
}

func (r *notificationOutboxRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *notificationOutboxRepository) Create(ctx context.Context, item *entity.NotificationOutbox) error {
	if err := r.getDB(ctx).Create(item).Error; err != nil {
		r.logger.Error("Failed to create notification outbox row", core.Error(err))
		return apperrors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *notificationOutboxRepository) ListPending(ctx context.Context, dueBefore time.Time, limit int) ([]*entity.NotificationOutbox, error) {
	if limit <= 0 {
		limit = 100
	}

	var rows []*entity.NotificationOutbox
	if err := r.getDB(ctx).
		Where("status = ?", entity.NotificationOutboxStatusPending).
		Where("next_attempt_at IS NULL OR next_attempt_at <= ?", dueBefore).
		Order("created_at ASC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		r.logger.Error("Failed to list pending outbox rows", core.Error(err))
		return nil, apperrors.InternalError("errors.databaseError", err)
	}

	return rows, nil
}

func (r *notificationOutboxRepository) MarkPublished(ctx context.Context, id uuid.UUID, publishedAt time.Time) error {
	updates := map[string]any{
		"status":       entity.NotificationOutboxStatusPublished,
		"published_at": publishedAt,
		"last_error":   nil,
	}

	if err := r.getDB(ctx).Model(&entity.NotificationOutbox{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		r.logger.Error("Failed to mark outbox row as published", core.Error(err))
		return apperrors.InternalError("errors.databaseError", err)
	}

	return nil
}

func (r *notificationOutboxRepository) MarkRetryScheduled(ctx context.Context, id uuid.UUID, attemptCount int, nextAttemptAt time.Time, lastError string) error {
	updates := map[string]any{
		"status":          entity.NotificationOutboxStatusPending,
		"attempt_count":   attemptCount,
		"next_attempt_at": nextAttemptAt,
		"last_error":      lastError,
	}

	if err := r.getDB(ctx).Model(&entity.NotificationOutbox{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		r.logger.Error("Failed to mark outbox row for retry", core.Error(err))
		return apperrors.InternalError("errors.databaseError", err)
	}

	return nil
}

func (r *notificationOutboxRepository) MarkDeadLetter(ctx context.Context, id uuid.UUID, attemptCount int, lastError string) error {
	updates := map[string]any{
		"status":        entity.NotificationOutboxStatusDeadLetter,
		"attempt_count": attemptCount,
		"last_error":    lastError,
	}

	if err := r.getDB(ctx).Model(&entity.NotificationOutbox{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		r.logger.Error("Failed to mark outbox row as dead-letter", core.Error(err))
		return apperrors.InternalError("errors.databaseError", err)
	}

	return nil
}
