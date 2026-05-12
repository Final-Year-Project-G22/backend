package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/domain/entity"
	"github.com/google/uuid"
)

// PaymentUseCase defines the payment and subscription business logic.
type PaymentUseCase interface {
	// InitiatePayment starts a new payment transaction.
	// Returns checkout URL and tx_ref.
	InitiatePayment(ctx context.Context, accountID uuid.UUID, email, firstName, lastName, phone, planName, period string) (*InitiatePaymentResult, error)

	// VerifyPayment verifies a payment by tx_ref.
	// Calls Chapa API and updates the payment record.
	VerifyPayment(ctx context.Context, accountID uuid.UUID, txRef string) (*VerifyPaymentResult, error)

	// HandleWebhook processes a Chapa webhook event.
	// Verifies signature, updates payment/subscription, emits notification.
	HandleWebhook(ctx context.Context, eventBody []byte, signature string) error

	// GetMySubscription returns the current account's active subscription.
	GetMySubscription(ctx context.Context, accountID uuid.UUID) (*entity.Subscription, error)

	// ListPlans returns available subscription plans.
	ListPlans(ctx context.Context) ([]*entity.Plan, error)
}

// InitiatePaymentResult is returned by InitiatePayment.
type InitiatePaymentResult struct {
	TxRef       string
	CheckoutURL string
	Amount      int64
	Currency    string
	PlanName    string
	Period      string
	ExpiresAt   int64 // Unix timestamp
}

// VerifyPaymentResult is returned by VerifyPayment.
type VerifyPaymentResult struct {
	TxRef         string
	ChapaRef      *string
	Status        entity.PaymentStatus
	Amount        int64
	Currency      string
	PlanName      string
	PlanPeriod    string
	PaymentMethod *string
	VerifiedAt    *string
}
