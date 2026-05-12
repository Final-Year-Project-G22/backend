package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
)

// PaymentHandler handles authenticated payment endpoints.
type PaymentHandler struct {
	usecase usecase.PaymentUseCase
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(uc usecase.PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{usecase: uc}
}

// HandleListPlans returns available subscription plans.
func (h *PaymentHandler) HandleListPlans(ctx context.Context, input *struct{}) (*dto.ListPlansOutput, error) {
	plans, err := h.usecase.ListPlans(ctx)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.ListPlansOutput{Body: dto.ListPlansResponseBody{
		Data: dto.ToPlanResponses(plans),
	}}, nil
}

// HandleInitiatePayment starts a new payment transaction.
func (h *PaymentHandler) HandleInitiatePayment(ctx context.Context, input *dto.InitiatePaymentInput) (*dto.InitiatePaymentOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	result, err := h.usecase.InitiatePayment(ctx, accountID, input.Body.Email, input.Body.FirstName, input.Body.LastName, input.Body.Phone, input.Body.PlanName, input.Body.Period)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.InitiatePaymentOutput{Body: dto.InitiatePaymentResponseBody{
		TxRef:       result.TxRef,
		CheckoutURL: result.CheckoutURL,
		Amount:      result.Amount,
		Currency:    result.Currency,
		PlanName:    result.PlanName,
		Period:      result.Period,
		ExpiresAt:   result.ExpiresAt,
	}}, nil
}

// HandleVerifyPayment verifies a payment by tx_ref.
func (h *PaymentHandler) HandleVerifyPayment(ctx context.Context, input *dto.VerifyPaymentInput) (*dto.VerifyPaymentOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	result, err := h.usecase.VerifyPayment(ctx, accountID, input.Body.TxRef)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.VerifyPaymentOutput{Body: dto.VerifyPaymentResponseBody{
		TxRef:         result.TxRef,
		ChapaRef:      strOrEmpty(result.ChapaRef),
		Status:        string(result.Status),
		Amount:        result.Amount,
		Currency:      result.Currency,
		PlanName:      result.PlanName,
		PlanPeriod:    result.PlanPeriod,
		PaymentMethod: strOrEmpty(result.PaymentMethod),
		VerifiedAt:    strOrEmpty(result.VerifiedAt),
	}}, nil
}

// HandleGetMySubscription returns the current account's active subscription.
func (h *PaymentHandler) HandleGetMySubscription(ctx context.Context, input *struct{}) (*dto.GetMySubscriptionOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	sub, err := h.usecase.GetMySubscription(ctx, accountID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.GetMySubscriptionOutput{Body: dto.GetMySubscriptionResponseBody{
		Data: dto.ToSubscriptionResponse(sub),
	}}, nil
}

// strOrEmpty returns the string value or empty string if nil.
func strOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
