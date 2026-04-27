package usecase

import (
	"context"

	"github.com/google/uuid"
)

type NotificationDeliveryUsecase interface {
	ProcessQueue(ctx context.Context, batchSize int) error
	DeliverItem(ctx context.Context, queueID uuid.UUID) error
	HandleDeliveryResult(ctx context.Context, queueID uuid.UUID, success bool, errMsg *string) error
	RetryFailed(ctx context.Context, batchSize int) error
	CancelPendingForAccount(ctx context.Context, accountID uuid.UUID) error
}
