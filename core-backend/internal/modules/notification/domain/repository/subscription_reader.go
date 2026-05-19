package repository

import (
	"context"

	"github.com/google/uuid"
)

type SubscriptionReader interface {
	HasActiveProSubscription(ctx context.Context, accountID uuid.UUID) (bool, error)
}
