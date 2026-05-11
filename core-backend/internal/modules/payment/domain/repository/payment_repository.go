package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

// PlanRepository defines operations for reading subscription plans.
type PlanRepository interface {
	sharedrepo.GenericRepository[entity.Plan]

	// FindActiveByNameAndPeriod returns an active plan by name and period.
	FindActiveByNameAndPeriod(ctx context.Context, name, period string) (*entity.Plan, error)

	// ListActive returns all active plans ordered by name and period.
	ListActive(ctx context.Context) ([]*entity.Plan, error)
}

// PaymentRepository defines operations for payment transactions.
type PaymentRepository interface {
	sharedrepo.GenericRepository[entity.Payment]

	// FindByTxRef returns a payment by its transaction reference.
	FindByTxRef(ctx context.Context, txRef string) (*entity.Payment, error)

	// FindPendingByAccount returns pending payments for an account.
	FindPendingByAccount(ctx context.Context, accountID uuid.UUID) ([]*entity.Payment, error)
}

// SubscriptionRepository defines operations for subscriptions.
type SubscriptionRepository interface {
	sharedrepo.GenericRepository[entity.Subscription]

	// GetActiveByAccount returns the active subscription for an account, if any.
	GetActiveByAccount(ctx context.Context, accountID uuid.UUID) (*entity.Subscription, error)

	// GetLatestByAccount returns the most recent subscription for an account.
	GetLatestByAccount(ctx context.Context, accountID uuid.UUID) (*entity.Subscription, error)
}
