package dto

import (
	"time"

	paymententity "github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/domain/entity"
	"github.com/google/uuid"
)

// --- List Plans ---

type ListPlansOutput struct {
	Body ListPlansResponseBody
}

type ListPlansResponseBody struct {
	Data []PlanResponse `json:"data" doc:"Available subscription plans"`
}

type PlanResponse struct {
	ID       uuid.UUID `json:"id" doc:"Plan ID"`
	Name     string    `json:"name" doc:"Plan name"`
	Period   string    `json:"period" doc:"Billing period"`
	Amount   int64     `json:"amount" doc:"Amount in minor units"`
	Currency string    `json:"currency" doc:"Currency code"`
	IsActive bool      `json:"isActive" doc:"Whether plan is available"`
}

// ToPlanResponse converts an entity.Plan to a PlanResponse.
func ToPlanResponse(plan *paymententity.Plan) PlanResponse {
	return PlanResponse{
		ID:       plan.ID,
		Name:     plan.Name,
		Period:   plan.Period,
		Amount:   plan.Amount,
		Currency: plan.Currency,
		IsActive: plan.IsActive,
	}
}

// ToPlanResponses converts a slice of entity.Plan to PlanResponse slice.
func ToPlanResponses(plans []*paymententity.Plan) []PlanResponse {
	result := make([]PlanResponse, len(plans))
	for i, plan := range plans {
		result[i] = ToPlanResponse(plan)
	}
	return result
}

// --- Initiate Payment ---

type InitiatePaymentInput struct {
	Body InitiatePaymentRequestBody
}

type InitiatePaymentRequestBody struct {
	PlanName string `json:"planName" doc:"Plan name (Basic or Pro)" example:"Pro"`
	Period   string `json:"period" doc:"Billing period (monthly or yearly)" example:"monthly"`
}

type InitiatePaymentOutput struct {
	Body InitiatePaymentResponseBody
}

type InitiatePaymentResponseBody struct {
	TxRef       string `json:"txRef" doc:"Transaction reference"`
	CheckoutURL string `json:"checkoutUrl" doc:"Chapa checkout URL"`
	Amount      int64  `json:"amount" doc:"Amount in minor units"`
	Currency    string `json:"currency" doc:"Currency code"`
	PlanName    string `json:"planName" doc:"Plan name"`
	Period      string `json:"period" doc:"Billing period"`
	ExpiresAt   int64  `json:"expiresAt" doc:"Checkout URL expiration (Unix timestamp)"`
}

// --- Verify Payment ---

type VerifyPaymentInput struct {
	Body VerifyPaymentRequestBody
}

type VerifyPaymentRequestBody struct {
	TxRef string `json:"txRef" doc:"Transaction reference to verify" example:"tx_abc123_Pro_monthly_20260510120000"`
}

type VerifyPaymentOutput struct {
	Body VerifyPaymentResponseBody
}

type VerifyPaymentResponseBody struct {
	TxRef         string `json:"txRef" doc:"Transaction reference"`
	ChapaRef      string `json:"chapaRef,omitempty" doc:"Chapa internal reference"`
	Status        string `json:"status" doc:"Payment status"`
	Amount        int64  `json:"amount" doc:"Amount in minor units"`
	Currency      string `json:"currency" doc:"Currency code"`
	PlanName      string `json:"planName" doc:"Plan name"`
	PlanPeriod    string `json:"planPeriod" doc:"Billing period"`
	PaymentMethod string `json:"paymentMethod,omitempty" doc:"Payment method used"`
	VerifiedAt    string `json:"verifiedAt,omitempty" doc:"Verification timestamp"`
}

// --- Get My Subscription ---

type GetMySubscriptionOutput struct {
	Body GetMySubscriptionResponseBody
}

type GetMySubscriptionResponseBody struct {
	Data *SubscriptionResponse `json:"data,omitempty" doc:"Current subscription, null if none"`
}

type SubscriptionResponse struct {
	ID                 uuid.UUID `json:"id" doc:"Subscription ID"`
	PlanName           string    `json:"planName" doc:"Plan name"`
	PlanPeriod         string    `json:"planPeriod" doc:"Billing period"`
	Amount             int64     `json:"amount" doc:"Amount in minor units"`
	Currency           string    `json:"currency" doc:"Currency code"`
	Status             string    `json:"status" doc:"Subscription status"`
	CurrentPeriodStart time.Time `json:"currentPeriodStart" doc:"Period start"`
	CurrentPeriodEnd   time.Time `json:"currentPeriodEnd" doc:"Period end"`
	RenewalCount       int       `json:"renewalCount" doc:"Number of renewals"`
}

// ToSubscriptionResponse converts an entity.Subscription to a SubscriptionResponse.
func ToSubscriptionResponse(sub *paymententity.Subscription) *SubscriptionResponse {
	if sub == nil {
		return nil
	}
	return &SubscriptionResponse{
		ID:                 sub.ID,
		PlanName:           sub.PlanName,
		PlanPeriod:         sub.PlanPeriod,
		Amount:             sub.Amount,
		Currency:           sub.Currency,
		Status:             string(sub.Status),
		CurrentPeriodStart: sub.CurrentPeriodStart,
		CurrentPeriodEnd:   sub.CurrentPeriodEnd,
		RenewalCount:       sub.RenewalCount,
	}
}
