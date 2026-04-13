package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/entity"
	"github.com/google/uuid"
)

type IngestionOutboxRepository interface {
	Create(ctx context.Context, item *entity.IngestionOutbox) error
	GetByEventID(ctx context.Context, eventID uuid.UUID) (*entity.IngestionOutbox, error)
	ListPending(ctx context.Context, dueBefore time.Time, limit int) ([]*entity.IngestionOutbox, error)
	MarkPublished(ctx context.Context, id uuid.UUID, publishedAt time.Time, signature []byte) error
	MarkFailed(ctx context.Context, id uuid.UUID, attemptCount int, nextAttemptAt time.Time, replayCount int32, lastError string) error
}
