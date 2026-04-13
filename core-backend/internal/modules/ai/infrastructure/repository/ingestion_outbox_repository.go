package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/entity"
	airepository "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/repository"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ingestionOutboxRepository struct {
	db     *core.Database
	logger core.Logger
}

func NewIngestionOutboxRepository(db *core.Database, logger core.Logger) airepository.IngestionOutboxRepository {
	return &ingestionOutboxRepository{db: db, logger: logger}
}

func (r *ingestionOutboxRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *ingestionOutboxRepository) Create(ctx context.Context, item *entity.IngestionOutbox) error {
	if err := r.getDB(ctx).Create(item).Error; err != nil {
		r.logger.Error("Failed to create ingestion outbox row", core.Error(err))
		return apperrors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *ingestionOutboxRepository) GetByEventID(ctx context.Context, eventID uuid.UUID) (*entity.IngestionOutbox, error) {
	var outbox entity.IngestionOutbox
	if err := r.getDB(ctx).Where("event_id = ?", eventID).First(&outbox).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Error("Failed to get outbox row by event ID", core.Error(err))
		return nil, apperrors.InternalError("errors.databaseError", err)
	}
	return &outbox, nil
}

func (r *ingestionOutboxRepository) ListPending(ctx context.Context, dueBefore time.Time, limit int) ([]*entity.IngestionOutbox, error) {
	if limit <= 0 {
		limit = 100
	}

	var rows []*entity.IngestionOutbox
	if err := r.getDB(ctx).
		Where("status = ?", entity.OutboxStatusPending).
		Where("next_attempt_at IS NULL OR next_attempt_at <= ?", dueBefore).
		Order("created_at ASC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		r.logger.Error("Failed to list pending outbox rows", core.Error(err))
		return nil, apperrors.InternalError("errors.databaseError", err)
	}

	return rows, nil
}

func (r *ingestionOutboxRepository) MarkPublished(ctx context.Context, id uuid.UUID, publishedAt time.Time, signature []byte) error {
	updates := map[string]any{
		"status":       entity.OutboxStatusPublished,
		"published_at": publishedAt,
		"signature":    signature,
		"last_error":   nil,
	}

	if err := r.getDB(ctx).Model(&entity.IngestionOutbox{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		r.logger.Error("Failed to mark outbox row as published", core.Error(err))
		return apperrors.InternalError("errors.databaseError", err)
	}

	return nil
}

func (r *ingestionOutboxRepository) MarkFailed(ctx context.Context, id uuid.UUID, attemptCount int, nextAttemptAt time.Time, replayCount int32, lastError string) error {
	updates := map[string]any{
		"status":          entity.OutboxStatusPending,
		"attempt_count":   attemptCount,
		"next_attempt_at": nextAttemptAt,
		"replay_count":    replayCount,
		"last_error":      lastError,
	}

	if err := r.getDB(ctx).Model(&entity.IngestionOutbox{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		r.logger.Error("Failed to mark outbox row as failed", core.Error(err))
		return apperrors.InternalError("errors.databaseError", err)
	}

	return nil
}
