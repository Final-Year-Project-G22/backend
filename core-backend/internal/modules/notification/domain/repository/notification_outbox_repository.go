package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/google/uuid"
)

type NotificationOutboxRepository interface {
	Create(ctx context.Context, item *entity.NotificationOutbox) error
	ListPending(ctx context.Context, dueBefore time.Time, limit int) ([]*entity.NotificationOutbox, error)
	MarkPublished(ctx context.Context, id uuid.UUID, publishedAt time.Time) error
	MarkRetryScheduled(ctx context.Context, id uuid.UUID, attemptCount int, nextAttemptAt time.Time, lastError string) error
	MarkDeadLetter(ctx context.Context, id uuid.UUID, attemptCount int, lastError string) error
}
