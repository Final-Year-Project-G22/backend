package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	notifentity "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/domain/entity"
	paymentrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/chapa"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// --- Mocks ---

type mockChapaClient struct {
	initResponse   *chapa.InitResponse
	initError      error
	verifyResponse *chapa.VerifyResponse
	verifyError    error
}

func (m *mockChapaClient) InitializeTransaction(ctx context.Context, req *chapa.InitRequest) (*chapa.InitResponse, error) {
	return m.initResponse, m.initError
}
func (m *mockChapaClient) VerifyTransaction(ctx context.Context, txRef string) (*chapa.VerifyResponse, error) {
	return m.verifyResponse, m.verifyError
}

type mockPlanRepo struct {
	plan  *entity.Plan
	plans []*entity.Plan
	err   error
}

func (m *mockPlanRepo) Create(ctx context.Context, item *entity.Plan) error        { return nil }
func (m *mockPlanRepo) BulkCreate(ctx context.Context, items []*entity.Plan) error { return nil }
func (m *mockPlanRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Plan, error) {
	return m.plan, m.err
}
func (m *mockPlanRepo) Update(ctx context.Context, item *entity.Plan) error { return nil }
func (m *mockPlanRepo) UpdateByID(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return nil
}
func (m *mockPlanRepo) Delete(ctx context.Context, id uuid.UUID) error     { return nil }
func (m *mockPlanRepo) HardDelete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockPlanRepo) FindAll(ctx context.Context, opts query.QueryOptions) sharedrepo.PaginatedResult[entity.Plan] {
	return sharedrepo.PaginatedResult[entity.Plan]{}
}
func (m *mockPlanRepo) FindAllArchived(ctx context.Context, opts query.QueryOptions) sharedrepo.PaginatedResult[entity.Plan] {
	return sharedrepo.PaginatedResult[entity.Plan]{}
}
func (m *mockPlanRepo) First(ctx context.Context, opts query.QueryOptions) (*entity.Plan, error) {
	return nil, nil
}
func (m *mockPlanRepo) Find(ctx context.Context, opts query.QueryOptions) ([]*entity.Plan, error) {
	return nil, nil
}
func (m *mockPlanRepo) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Plan, error) {
	return nil, nil
}
func (m *mockPlanRepo) Exists(ctx context.Context, id uuid.UUID) (bool, error) { return false, nil }
func (m *mockPlanRepo) Count(ctx context.Context) (int64, error)               { return 0, nil }
func (m *mockPlanRepo) Transaction(ctx context.Context, fn func(repo sharedrepo.GenericRepository[entity.Plan]) error) error {
	return nil
}
func (m *mockPlanRepo) GetDB() *gorm.DB { return nil }
func (m *mockPlanRepo) FindActiveByNameAndPeriod(ctx context.Context, name, period string) (*entity.Plan, error) {
	return m.plan, m.err
}
func (m *mockPlanRepo) ListActive(ctx context.Context) ([]*entity.Plan, error) {
	return m.plans, m.err
}

type mockPaymentRepo struct {
	payment  *entity.Payment
	payments []*entity.Payment
	err      error
}

func (m *mockPaymentRepo) Create(ctx context.Context, item *entity.Payment) error        { return nil }
func (m *mockPaymentRepo) BulkCreate(ctx context.Context, items []*entity.Payment) error { return nil }
func (m *mockPaymentRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Payment, error) {
	return m.payment, m.err
}
func (m *mockPaymentRepo) Update(ctx context.Context, item *entity.Payment) error { return nil }
func (m *mockPaymentRepo) UpdateByID(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return nil
}
func (m *mockPaymentRepo) Delete(ctx context.Context, id uuid.UUID) error     { return nil }
func (m *mockPaymentRepo) HardDelete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockPaymentRepo) FindAll(ctx context.Context, opts query.QueryOptions) sharedrepo.PaginatedResult[entity.Payment] {
	return sharedrepo.PaginatedResult[entity.Payment]{}
}
func (m *mockPaymentRepo) FindAllArchived(ctx context.Context, opts query.QueryOptions) sharedrepo.PaginatedResult[entity.Payment] {
	return sharedrepo.PaginatedResult[entity.Payment]{}
}
func (m *mockPaymentRepo) First(ctx context.Context, opts query.QueryOptions) (*entity.Payment, error) {
	return nil, nil
}
func (m *mockPaymentRepo) Find(ctx context.Context, opts query.QueryOptions) ([]*entity.Payment, error) {
	return nil, nil
}
func (m *mockPaymentRepo) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Payment, error) {
	return nil, nil
}
func (m *mockPaymentRepo) Exists(ctx context.Context, id uuid.UUID) (bool, error) { return false, nil }
func (m *mockPaymentRepo) Count(ctx context.Context) (int64, error)               { return 0, nil }
func (m *mockPaymentRepo) Transaction(ctx context.Context, fn func(repo sharedrepo.GenericRepository[entity.Payment]) error) error {
	return nil
}
func (m *mockPaymentRepo) GetDB() *gorm.DB { return nil }
func (m *mockPaymentRepo) FindByTxRef(ctx context.Context, txRef string) (*entity.Payment, error) {
	return m.payment, m.err
}
func (m *mockPaymentRepo) FindPendingByAccount(ctx context.Context, accountID uuid.UUID) ([]*entity.Payment, error) {
	return m.payments, m.err
}

type mockSubscriptionRepo struct {
	sub *entity.Subscription
	err error
}

func (m *mockSubscriptionRepo) Create(ctx context.Context, item *entity.Subscription) error {
	return nil
}
func (m *mockSubscriptionRepo) BulkCreate(ctx context.Context, items []*entity.Subscription) error {
	return nil
}
func (m *mockSubscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error) {
	return m.sub, m.err
}
func (m *mockSubscriptionRepo) Update(ctx context.Context, item *entity.Subscription) error {
	return nil
}
func (m *mockSubscriptionRepo) UpdateByID(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return nil
}
func (m *mockSubscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error     { return nil }
func (m *mockSubscriptionRepo) HardDelete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockSubscriptionRepo) FindAll(ctx context.Context, opts query.QueryOptions) sharedrepo.PaginatedResult[entity.Subscription] {
	return sharedrepo.PaginatedResult[entity.Subscription]{}
}
func (m *mockSubscriptionRepo) FindAllArchived(ctx context.Context, opts query.QueryOptions) sharedrepo.PaginatedResult[entity.Subscription] {
	return sharedrepo.PaginatedResult[entity.Subscription]{}
}
func (m *mockSubscriptionRepo) First(ctx context.Context, opts query.QueryOptions) (*entity.Subscription, error) {
	return nil, nil
}
func (m *mockSubscriptionRepo) Find(ctx context.Context, opts query.QueryOptions) ([]*entity.Subscription, error) {
	return nil, nil
}
func (m *mockSubscriptionRepo) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Subscription, error) {
	return nil, nil
}
func (m *mockSubscriptionRepo) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockSubscriptionRepo) Count(ctx context.Context) (int64, error) { return 0, nil }
func (m *mockSubscriptionRepo) Transaction(ctx context.Context, fn func(repo sharedrepo.GenericRepository[entity.Subscription]) error) error {
	return nil
}
func (m *mockSubscriptionRepo) GetDB() *gorm.DB { return nil }
func (m *mockSubscriptionRepo) GetActiveByAccount(ctx context.Context, accountID uuid.UUID) (*entity.Subscription, error) {
	return m.sub, m.err
}
func (m *mockSubscriptionRepo) GetLatestByAccount(ctx context.Context, accountID uuid.UUID) (*entity.Subscription, error) {
	return m.sub, m.err
}

type mockNotificationOutboxRepo struct {
	created bool
	err     error
}

func (m *mockNotificationOutboxRepo) Create(ctx context.Context, item *notifentity.NotificationOutbox) error {
	m.created = true
	return m.err
}
func (m *mockNotificationOutboxRepo) ListPending(ctx context.Context, dueBefore time.Time, limit int) ([]*notifentity.NotificationOutbox, error) {
	return nil, nil
}
func (m *mockNotificationOutboxRepo) MarkPublished(ctx context.Context, id uuid.UUID, publishedAt time.Time) error {
	return nil
}
func (m *mockNotificationOutboxRepo) MarkRetryScheduled(ctx context.Context, id uuid.UUID, attemptCount int, nextAttemptAt time.Time, reason string) error {
	return nil
}
func (m *mockNotificationOutboxRepo) MarkDeadLetter(ctx context.Context, id uuid.UUID, attemptCount int, reason string) error {
	return nil
}

type mockTransactor struct {
	fn func(ctx context.Context) error
}

func (m *mockTransactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.fn != nil {
		return m.fn(ctx)
	}
	return fn(ctx)
}

type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields ...zap.Field)       {}
func (m *mockLogger) Info(msg string, fields ...zap.Field)        {}
func (m *mockLogger) Warn(msg string, fields ...zap.Field)        {}
func (m *mockLogger) Error(msg string, fields ...zap.Field)       {}
func (m *mockLogger) Fatal(msg string, fields ...zap.Field)       {}
func (m *mockLogger) With(fields ...zap.Field) core.Logger        { return m }
func (m *mockLogger) WithContext(ctx context.Context) core.Logger { return m }
func (m *mockLogger) Sync() error                                 { return nil }

// --- Tests ---

func TestPaymentUseCase_ListPlans(t *testing.T) {
	plans := []*entity.Plan{
		{Name: "Pro", Period: "monthly", Amount: 19900, Currency: "ETB", IsActive: true},
	}
	uc := NewPaymentUseCase(
		nil,
		&mockPlanRepo{plans: plans},
		&mockPaymentRepo{},
		&mockSubscriptionRepo{},
		&mockNotificationOutboxRepo{},
		&core.Config{},
		&mockTransactor{},
		&mockLogger{},
	)

	result, err := uc.ListPlans(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 plan, got %d", len(result))
	}
	if result[0].Name != "Pro" {
		t.Errorf("expected Pro, got %s", result[0].Name)
	}
}

func TestPaymentUseCase_GetMySubscription_Active(t *testing.T) {
	accountID := uuid.New()
	sub := &entity.Subscription{
		AccountID: accountID,
		PlanName:  "Pro",
		Status:    entity.SubscriptionStatusActive,
	}
	uc := NewPaymentUseCase(
		nil, nil, nil,
		&mockSubscriptionRepo{sub: sub},
		&mockNotificationOutboxRepo{},
		&core.Config{},
		&mockTransactor{},
		&mockLogger{},
	)

	result, err := uc.GetMySubscription(context.Background(), accountID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected subscription, got nil")
	}
	if result.PlanName != "Pro" {
		t.Errorf("expected Pro, got %s", result.PlanName)
	}
}

func TestPaymentUseCase_GetMySubscription_None(t *testing.T) {
	accountID := uuid.New()
	uc := NewPaymentUseCase(
		nil, nil, nil,
		&mockSubscriptionRepo{sub: nil},
		&mockNotificationOutboxRepo{},
		&core.Config{},
		&mockTransactor{},
		&mockLogger{},
	)

	result, err := uc.GetMySubscription(context.Background(), accountID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil subscription")
	}
}

func TestPaymentUseCase_InitiatePayment_Success(t *testing.T) {
	accountID := uuid.New()
	plan := &entity.Plan{Name: "Pro", Period: "monthly", Amount: 19900, Currency: "ETB", IsActive: true}

	chapaClient := &mockChapaClient{
		initResponse: &chapa.InitResponse{
			Status: "success",
			Data:   chapa.InitResponseData{CheckoutURL: "https://checkout.chapa.co/test"},
		},
	}

	uc := NewPaymentUseCase(
		chapaClient,
		&mockPlanRepo{plan: plan},
		&mockPaymentRepo{payment: nil},
		&mockSubscriptionRepo{},
		&mockNotificationOutboxRepo{},
		&core.Config{Chapa: core.ChapaConfig{CallbackURL: "https://api.example.com/webhooks/chapa", ReturnURL: "adisu://payment/success"}},
		&mockTransactor{},
		&mockLogger{},
	)

	result, err := uc.InitiatePayment(context.Background(), accountID, "Pro", "monthly")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CheckoutURL != "https://checkout.chapa.co/test" {
		t.Errorf("expected checkout URL, got %s", result.CheckoutURL)
	}
	if result.Amount != 19900 {
		t.Errorf("expected amount 19900, got %d", result.Amount)
	}
}

func TestPaymentUseCase_InitiatePayment_PlanNotFound(t *testing.T) {
	accountID := uuid.New()
	uc := NewPaymentUseCase(
		nil,
		&mockPlanRepo{plan: nil},
		&mockPaymentRepo{},
		&mockSubscriptionRepo{},
		&mockNotificationOutboxRepo{},
		&core.Config{},
		&mockTransactor{},
		&mockLogger{},
	)

	_, err := uc.InitiatePayment(context.Background(), accountID, "NonExistent", "monthly")
	if err == nil {
		t.Fatal("expected error for missing plan")
	}
	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("expected *errors.AppError, got %T", err)
	}
	if appErr.GetStatus() != 400 {
		t.Errorf("expected status 400, got %d", appErr.GetStatus())
	}
}

func TestPaymentUseCase_InitiatePayment_AlreadyPaid(t *testing.T) {
	accountID := uuid.New()
	plan := &entity.Plan{Name: "Pro", Period: "monthly", Amount: 19900, Currency: "ETB", IsActive: true}
	existingPayment := &entity.Payment{
		AccountID: accountID,
		TxRef:     generateTxRef(accountID, "Pro", "monthly"),
		Status:    entity.PaymentStatusSuccess,
	}

	uc := NewPaymentUseCase(
		nil,
		&mockPlanRepo{plan: plan},
		&mockPaymentRepo{payment: existingPayment},
		&mockSubscriptionRepo{},
		&mockNotificationOutboxRepo{},
		&core.Config{},
		&mockTransactor{},
		&mockLogger{},
	)

	_, err := uc.InitiatePayment(context.Background(), accountID, "Pro", "monthly")
	if err == nil {
		t.Fatal("expected error for already paid")
	}
}

func TestPaymentUseCase_VerifyPayment_AlreadyTerminal(t *testing.T) {
	accountID := uuid.New()
	payment := &entity.Payment{
		AccountID: accountID,
		TxRef:     "tx_test",
		Status:    entity.PaymentStatusSuccess,
		Amount:    19900,
		Currency:  "ETB",
	}

	uc := NewPaymentUseCase(
		nil, nil,
		&mockPaymentRepo{payment: payment},
		&mockSubscriptionRepo{},
		&mockNotificationOutboxRepo{},
		&core.Config{},
		&mockTransactor{},
		&mockLogger{},
	)

	result, err := uc.VerifyPayment(context.Background(), accountID, "tx_test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != entity.PaymentStatusSuccess {
		t.Errorf("expected success, got %s", result.Status)
	}
}

func TestPaymentUseCase_VerifyPayment_NotFound(t *testing.T) {
	accountID := uuid.New()
	uc := NewPaymentUseCase(
		nil, nil,
		&mockPaymentRepo{payment: nil},
		&mockSubscriptionRepo{},
		&mockNotificationOutboxRepo{},
		&core.Config{},
		&mockTransactor{},
		&mockLogger{},
	)

	_, err := uc.VerifyPayment(context.Background(), accountID, "tx_nonexistent")
	if err == nil {
		t.Fatal("expected error for missing payment")
	}
}

func TestPaymentUseCase_HandleWebhook_SignatureInvalid(t *testing.T) {
	uc := NewPaymentUseCase(
		nil, nil, nil, nil,
		&mockNotificationOutboxRepo{},
		&core.Config{Chapa: core.ChapaConfig{WebhookSecret: "secret"}},
		&mockTransactor{},
		&mockLogger{},
	)

	err := uc.HandleWebhook(context.Background(), []byte(`{"event":"charge.success"}`), "invalid_sig")
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}
}

func TestPaymentUseCase_HandleWebhook_UnknownPayment(t *testing.T) {
	uc := NewPaymentUseCase(
		nil, nil,
		&mockPaymentRepo{payment: nil},
		&mockSubscriptionRepo{},
		&mockNotificationOutboxRepo{},
		&core.Config{Chapa: core.ChapaConfig{WebhookSecret: "secret"}},
		&mockTransactor{},
		&mockLogger{},
	)

	body := []byte(`{"event":"charge.success","tx_ref":"tx_unknown"}`)
	sig := generateValidSignature(body, "secret")

	err := uc.HandleWebhook(context.Background(), body, sig)
	if err != nil {
		t.Fatalf("expected nil for unknown payment (acknowledge), got: %v", err)
	}
}

func TestPaymentUseCase_HandleWebhook_Idempotent(t *testing.T) {
	accountID := uuid.New()
	payment := &entity.Payment{
		AccountID: accountID,
		TxRef:     "tx_test",
		Status:    entity.PaymentStatusSuccess,
	}
	uc := NewPaymentUseCase(
		nil, nil,
		&mockPaymentRepo{payment: payment},
		&mockSubscriptionRepo{},
		&mockNotificationOutboxRepo{},
		&core.Config{Chapa: core.ChapaConfig{WebhookSecret: "secret"}},
		&mockTransactor{},
		&mockLogger{},
	)

	body := []byte(`{"event":"charge.success","tx_ref":"tx_test"}`)
	sig := generateValidSignature(body, "secret")

	err := uc.HandleWebhook(context.Background(), body, sig)
	if err != nil {
		t.Fatalf("expected nil for already processed, got: %v", err)
	}
}

func TestGenerateTxRef(t *testing.T) {
	accountID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	txRef := generateTxRef(accountID, "Pro", "monthly")
	if txRef == "" {
		t.Fatal("expected non-empty tx_ref")
	}
	if len(txRef) < 20 {
		t.Errorf("expected tx_ref to be longer than 20 chars, got %d", len(txRef))
	}
}

func TestExtractCheckoutURL(t *testing.T) {
	payment := &entity.Payment{
		Metadata: datatypes.JSONMap{"checkout_url": "https://checkout.chapa.co/test"},
	}
	url := extractCheckoutURL(payment)
	if url != "https://checkout.chapa.co/test" {
		t.Errorf("expected checkout URL, got %s", url)
	}

	payment.Metadata = nil
	url = extractCheckoutURL(payment)
	if url != "" {
		t.Errorf("expected empty string for nil metadata, got %s", url)
	}
}

// --- Interface assertions ---

var _ chapa.Client = (*mockChapaClient)(nil)
var _ paymentrepo.PlanRepository = (*mockPlanRepo)(nil)
var _ paymentrepo.PaymentRepository = (*mockPaymentRepo)(nil)
var _ paymentrepo.SubscriptionRepository = (*mockSubscriptionRepo)(nil)
var _ notifrepo.NotificationOutboxRepository = (*mockNotificationOutboxRepo)(nil)
var _ sharedrepo.Transactor = (*mockTransactor)(nil)
var _ core.Logger = (*mockLogger)(nil)

func generateValidSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
