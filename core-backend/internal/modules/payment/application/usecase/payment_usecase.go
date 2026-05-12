package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	notifentity "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notifevent "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/event"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/domain/entity"
	paymentrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/domain/repository"
	paymentuc "github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/notificationevent"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/chapa"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// paymentUseCase implements PaymentUseCase.
type paymentUseCase struct {
	chapaClient      chapa.Client
	planRepo         paymentrepo.PlanRepository
	paymentRepo      paymentrepo.PaymentRepository
	subscriptionRepo paymentrepo.SubscriptionRepository
	notifOutboxRepo  notifrepo.NotificationOutboxRepository
	cfg              *core.Config
	transactor       sharedrepo.Transactor
	logger           core.Logger
}

// NewPaymentUseCase creates a new payment usecase.
func NewPaymentUseCase(
	chapaClient chapa.Client,
	planRepo paymentrepo.PlanRepository,
	paymentRepo paymentrepo.PaymentRepository,
	subscriptionRepo paymentrepo.SubscriptionRepository,
	notifOutboxRepo notifrepo.NotificationOutboxRepository,
	cfg *core.Config,
	transactor sharedrepo.Transactor,
	logger core.Logger,
) paymentuc.PaymentUseCase {
	return &paymentUseCase{
		chapaClient:      chapaClient,
		planRepo:         planRepo,
		paymentRepo:      paymentRepo,
		subscriptionRepo: subscriptionRepo,
		notifOutboxRepo:  notifOutboxRepo,
		cfg:              cfg,
		transactor:       transactor,
		logger:           logger,
	}
}

// ListPlans returns all active subscription plans.
func (u *paymentUseCase) ListPlans(ctx context.Context) ([]*entity.Plan, error) {
	return u.planRepo.ListActive(ctx)
}

// GetMySubscription returns the current account's active subscription.
func (u *paymentUseCase) GetMySubscription(ctx context.Context, accountID uuid.UUID) (*entity.Subscription, error) {
	return u.subscriptionRepo.GetActiveByAccount(ctx, accountID)
}

// InitiatePayment starts a new payment transaction.
func (u *paymentUseCase) InitiatePayment(ctx context.Context, accountID uuid.UUID, email, firstName, lastName, phone, planName, period string) (*paymentuc.InitiatePaymentResult, error) {
	// 1. Validate plan exists and is active
	plan, err := u.planRepo.FindActiveByNameAndPeriod(ctx, planName, period)
	if err != nil {
		return nil, errors.InternalError("errors.databaseError", err)
	}
	if plan == nil {
		return nil, errors.BadRequestError("errors.planNotFound")
	}

	// 2. Generate deterministic tx_ref
	txRef := generateTxRef(accountID, planName, period)

	// 3. Check for duplicate pending payment
	existing, err := u.paymentRepo.FindByTxRef(ctx, txRef)
	if err != nil {
		return nil, errors.InternalError("errors.databaseError", err)
	}
	if existing != nil && existing.Status == entity.PaymentStatusPending {
		// Return existing checkout URL if still pending and not expired
		if existing.CheckedOutAt != nil && time.Since(*existing.CheckedOutAt) < 30*time.Minute {
			return &paymentuc.InitiatePaymentResult{
				TxRef:       existing.TxRef,
				CheckoutURL: extractCheckoutURL(existing),
				Amount:      existing.Amount,
				Currency:    existing.Currency,
				PlanName:    existing.PlanName,
				Period:      existing.PlanPeriod,
				ExpiresAt:   existing.CheckedOutAt.Add(30 * time.Minute).Unix(),
			}, nil
		}
	}
	if existing != nil && existing.Status == entity.PaymentStatusSuccess {
		return nil, errors.BadRequestError("errors.alreadyPaid")
	}

	// 4. Create payment record
	now := time.Now().UTC()
	payment := &entity.Payment{
		AccountID:    accountID,
		TxRef:        txRef,
		Amount:       plan.Amount,
		Currency:     plan.Currency,
		PlanName:     plan.Name,
		PlanPeriod:   plan.Period,
		Status:       entity.PaymentStatusPending,
		CheckedOutAt: &now,
		Metadata:     datatypes.JSONMap{},
	}
	if err := u.paymentRepo.Create(ctx, payment); err != nil {
		return nil, errors.InternalError("errors.databaseError", err)
	}

	// 5. Call Chapa to initialize transaction
	chapaReq := &chapa.InitRequest{
		Amount:      plan.Amount,
		Currency:    plan.Currency,
		Email:       email,
		FirstName:   firstName,
		LastName:    lastName,
		Phone:       phone,
		TxRef:       txRef,
		CallbackURL: u.cfg.Chapa.CallbackURL,
		ReturnURL:   u.cfg.Chapa.ReturnURL,
		Customization: map[string]interface{}{
			"title":       "Adisu Pro",
			"description": fmt.Sprintf("%s %s plan", plan.Name, plan.Period),
		},
	}

	chapaResp, err := u.chapaClient.InitializeTransaction(ctx, chapaReq)
	if err != nil {
		u.logger.Error("chapa initialize failed", core.Error(err), core.String("txRef", txRef))
		return nil, errors.InternalError("errors.paymentInitFailed", err)
	}

	// 6. Update payment with Chapa reference
	if chapaResp.Data.CheckoutURL != "" {
		payment.Metadata["checkout_url"] = chapaResp.Data.CheckoutURL
		if err := u.paymentRepo.Update(ctx, payment); err != nil {
			u.logger.Error("failed to update payment metadata", core.Error(err))
		}
	}

	return &paymentuc.InitiatePaymentResult{
		TxRef:       txRef,
		CheckoutURL: chapaResp.Data.CheckoutURL,
		Amount:      plan.Amount,
		Currency:    plan.Currency,
		PlanName:    plan.Name,
		Period:      plan.Period,
		ExpiresAt:   now.Add(30 * time.Minute).Unix(),
	}, nil
}

// VerifyPayment verifies a payment by tx_ref and updates records.
func (u *paymentUseCase) VerifyPayment(ctx context.Context, accountID uuid.UUID, txRef string) (*paymentuc.VerifyPaymentResult, error) {
	payment, err := u.paymentRepo.FindByTxRef(ctx, txRef)
	if err != nil {
		return nil, errors.InternalError("errors.databaseError", err)
	}
	if payment == nil {
		return nil, errors.NotFoundError("payment", txRef)
	}
	if payment.AccountID != accountID {
		return nil, errors.ForbiddenError("errors.unauthorized")
	}

	// If already terminal, return current state
	if payment.Status == entity.PaymentStatusSuccess || payment.Status == entity.PaymentStatusFailed {
		return u.toVerifyResult(payment), nil
	}

	// Call Chapa verify
	return u.processVerification(ctx, payment)
}

// HandleWebhook processes a Chapa webhook event.
func (u *paymentUseCase) HandleWebhook(ctx context.Context, eventBody []byte, signature string) error {
	// 1. Verify signature
	if !chapa.VerifySignature(eventBody, signature, u.cfg.Chapa.WebhookSecret) {
		return errors.UnauthorizedError("errors.invalidSignature")
	}

	// 2. Parse event
	var event chapa.WebhookEvent
	if err := json.Unmarshal(eventBody, &event); err != nil {
		return errors.BadRequestError("errors.invalidWebhookPayload")
	}

	// 3. Find payment
	payment, err := u.paymentRepo.FindByTxRef(ctx, event.TxRef)
	if err != nil {
		return errors.InternalError("errors.databaseError", err)
	}
	if payment == nil {
		u.logger.Warn("webhook received for unknown payment", core.String("txRef", event.TxRef))
		return nil // Acknowledge to stop retries
	}

	// 4. Idempotency check
	if payment.Status == entity.PaymentStatusSuccess || payment.Status == entity.PaymentStatusFailed {
		return nil // Already processed
	}

	// 5. Re-verify with Chapa API (best practice)
	_, err = u.processVerification(ctx, payment)
	if err != nil {
		u.logger.Error("webhook verification failed", core.Error(err), core.String("txRef", event.TxRef))
		// Still return nil to acknowledge webhook
		return nil
	}

	return nil
}

// processVerification calls Chapa verify API and updates payment + subscription atomically.
func (u *paymentUseCase) processVerification(ctx context.Context, payment *entity.Payment) (*paymentuc.VerifyPaymentResult, error) {
	chapaResp, err := u.chapaClient.VerifyTransaction(ctx, payment.TxRef)
	if err != nil {
		u.logger.Error("chapa verify failed", core.Error(err), core.String("txRef", payment.TxRef))
		return nil, errors.InternalError("errors.paymentVerifyFailed", err)
	}

	// Determine status from Chapa response
	var newStatus entity.PaymentStatus
	var verifiedAt, failedAt *time.Time
	now := time.Now().UTC()

	switch chapaResp.Data.Status {
	case "success":
		newStatus = entity.PaymentStatusSuccess
		verifiedAt = &now
	case "failed", "cancelled":
		newStatus = entity.PaymentStatusFailed
		failedAt = &now
	default:
		// Still pending — don't update
		return u.toVerifyResult(payment), nil
	}

	// Update payment and subscription in transaction
	var result *paymentuc.VerifyPaymentResult
	err = u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Re-read payment inside transaction for idempotency
		p, err := u.paymentRepo.FindByTxRef(txCtx, payment.TxRef)
		if err != nil {
			return err
		}
		if p == nil {
			return fmt.Errorf("payment not found in transaction")
		}
		if p.Status == entity.PaymentStatusSuccess || p.Status == entity.PaymentStatusFailed {
			return nil // Already processed by concurrent request
		}

		// Update payment
		p.Status = newStatus
		p.ChapaRef = &chapaResp.Data.Reference
		p.VerifiedAt = verifiedAt
		p.FailedAt = failedAt
		p.PaymentMethod = &chapaResp.Data.PaymentMethod
		if p.Metadata == nil {
			p.Metadata = datatypes.JSONMap{}
		}
		p.Metadata["verify_response"] = chapaResp

		if err := u.paymentRepo.Update(txCtx, p); err != nil {
			return err
		}

		// If success, create/update subscription
		if newStatus == entity.PaymentStatusSuccess {
			if err := u.upsertSubscription(txCtx, p); err != nil {
				return err
			}
			if err := u.writePaymentConfirmationOutbox(txCtx, p); err != nil {
				u.logger.Error("failed to write payment confirmation outbox", core.Error(err))
				// Don't fail the transaction
			}
		}

		result = u.toVerifyResult(p)
		return nil
	})

	if err != nil {
		return nil, errors.InternalError("errors.transactionFailed", err)
	}

	return result, nil
}

// upsertSubscription creates or updates a subscription for a successful payment.
func (u *paymentUseCase) upsertSubscription(ctx context.Context, payment *entity.Payment) error {
	existing, err := u.subscriptionRepo.GetActiveByAccount(ctx, payment.AccountID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	var periodEnd time.Time
	if payment.PlanPeriod == "yearly" {
		periodEnd = now.AddDate(1, 0, 0)
	} else {
		periodEnd = now.AddDate(0, 1, 0)
	}

	if existing != nil {
		// Update existing subscription
		existing.PlanName = payment.PlanName
		existing.PlanPeriod = payment.PlanPeriod
		existing.Amount = payment.Amount
		existing.Currency = payment.Currency
		existing.Status = entity.SubscriptionStatusActive
		existing.CurrentPeriodStart = now
		existing.CurrentPeriodEnd = periodEnd
		existing.RenewalCount++
		if err := u.subscriptionRepo.Update(ctx, existing); err != nil {
			return err
		}
		// Link payment to subscription
		payment.SubscriptionID = &existing.ID
		return u.paymentRepo.Update(ctx, payment)
	}

	// Create new subscription
	sub := &entity.Subscription{
		AccountID:          payment.AccountID,
		PlanName:           payment.PlanName,
		PlanPeriod:         payment.PlanPeriod,
		Amount:             payment.Amount,
		Currency:           payment.Currency,
		Status:             entity.SubscriptionStatusActive,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   periodEnd,
		RenewalCount:       1,
	}
	if err := u.subscriptionRepo.Create(ctx, sub); err != nil {
		return err
	}

	// Link payment to subscription
	payment.SubscriptionID = &sub.ID
	return u.paymentRepo.Update(ctx, payment)
}

// writePaymentConfirmationOutbox writes a notification outbox entry for payment success.
func (u *paymentUseCase) writePaymentConfirmationOutbox(ctx context.Context, payment *entity.Payment) error {
	env := notificationevent.Envelope{
		SchemaVersion:    notificationevent.SchemaVersionV1,
		EventType:        notifevent.PaymentConfirmation,
		OccurredAt:       time.Now().UTC(),
		SourceModule:     "payment",
		AccountID:        payment.AccountID,
		NotificationType: string(notifentity.NotificationTypePaymentConfirmation),
		ChannelPolicy:    notificationevent.ChannelPolicyAllEnabled,
		Variables: map[string]string{
			"planName": payment.PlanName,
			"period":   payment.PlanPeriod,
			"amount":   strconv.FormatInt(payment.Amount, 10),
			"txRef":    payment.TxRef,
		},
		Metadata: notificationevent.Metadata{
			IdempotencyKey: fmt.Sprintf("payment:%s", payment.TxRef),
		},
	}

	data, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("failed to marshal envelope: %w", err)
	}

	var payload datatypes.JSONMap
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("failed to convert envelope to JSONMap: %w", err)
	}

	outbox := &notifentity.NotificationOutbox{
		EventType:      env.EventType,
		SchemaVersion:  env.SchemaVersion,
		SourceModule:   env.SourceModule,
		AccountID:      env.AccountID,
		IdempotencyKey: env.Metadata.IdempotencyKey,
		Payload:        payload,
		Status:         notifentity.NotificationOutboxStatusPending,
		AttemptCount:   0,
	}

	return u.notifOutboxRepo.Create(ctx, outbox)
}

// toVerifyResult converts a Payment entity to VerifyPaymentResult.
func (u *paymentUseCase) toVerifyResult(payment *entity.Payment) *paymentuc.VerifyPaymentResult {
	var verifiedAtStr *string
	if payment.VerifiedAt != nil {
		s := payment.VerifiedAt.Format(time.RFC3339)
		verifiedAtStr = &s
	}
	return &paymentuc.VerifyPaymentResult{
		TxRef:         payment.TxRef,
		ChapaRef:      payment.ChapaRef,
		Status:        payment.Status,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		PlanName:      payment.PlanName,
		PlanPeriod:    payment.PlanPeriod,
		PaymentMethod: payment.PaymentMethod,
		VerifiedAt:    verifiedAtStr,
	}
}

// generateTxRef creates a deterministic transaction reference.
func generateTxRef(accountID uuid.UUID, planName, period string) string {
	timestamp := time.Now().UTC().Format("20060102150405")
	return fmt.Sprintf("tx_%s_%s_%s_%s", accountID.String()[:8], planName, period, timestamp)
}

// extractCheckoutURL extracts the checkout URL from payment metadata.
func extractCheckoutURL(payment *entity.Payment) string {
	if payment.Metadata == nil {
		return ""
	}
	if url, ok := payment.Metadata["checkout_url"].(string); ok {
		return url
	}
	return ""
}
