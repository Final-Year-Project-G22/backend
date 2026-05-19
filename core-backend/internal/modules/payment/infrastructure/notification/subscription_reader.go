package notification

import (
	"context"

	paymentrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/domain/repository"
	"github.com/google/uuid"
)

type SubscriptionReaderAdapter struct {
	subscriptionRepo paymentrepo.SubscriptionRepository
}

func NewSubscriptionReaderAdapter(subscriptionRepo paymentrepo.SubscriptionRepository) *SubscriptionReaderAdapter {
	return &SubscriptionReaderAdapter{subscriptionRepo: subscriptionRepo}
}

func (a *SubscriptionReaderAdapter) HasActiveProSubscription(ctx context.Context, accountID uuid.UUID) (bool, error) {
	sub, err := a.subscriptionRepo.GetActiveByAccount(ctx, accountID)
	if err != nil {
		return false, err
	}
	if sub == nil {
		return false, nil
	}
	return sub.PlanName == "Pro", nil
}
