package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type NotificationQueueRepository interface {
	sharedrepo.GenericRepository[entity.NotificationQueue]

	FetchPending(ctx context.Context, limit int) ([]*entity.NotificationQueue, error)
	MarkProcessing(ctx context.Context, id uuid.UUID) error
	MarkDelivered(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error
	IncrementRetry(ctx context.Context, id uuid.UUID, nextScheduledFor time.Time) error
	CancelByAccount(ctx context.Context, accountID uuid.UUID) error
	CancelByCampaign(ctx context.Context, campaignID uuid.UUID) error
	CountByStatus(ctx context.Context, status entity.NotificationStatus) (int64, error)
}
