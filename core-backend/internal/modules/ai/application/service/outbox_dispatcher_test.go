package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/entity"
	aisvc "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/service"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type fakeOutboxRepo struct {
	rows            []*entity.IngestionOutbox
	markPublished   int
	markRetry       int
	markDead        int
	lastRetryID     uuid.UUID
	lastDeadID      uuid.UUID
	lastRetryReason string
	lastDeadReason  string
}

func (f *fakeOutboxRepo) Create(context.Context, *entity.IngestionOutbox) error { return nil }
func (f *fakeOutboxRepo) GetByEventID(context.Context, uuid.UUID) (*entity.IngestionOutbox, error) {
	return nil, nil
}
func (f *fakeOutboxRepo) ListPending(context.Context, time.Time, int) ([]*entity.IngestionOutbox, error) {
	return f.rows, nil
}
func (f *fakeOutboxRepo) MarkPublished(context.Context, uuid.UUID, time.Time, []byte) error {
	f.markPublished++
	return nil
}
func (f *fakeOutboxRepo) MarkRetryScheduled(_ context.Context, id uuid.UUID, _ int, _ time.Time, _ int32, reason string) error {
	f.markRetry++
	f.lastRetryID = id
	f.lastRetryReason = reason
	return nil
}
func (f *fakeOutboxRepo) MarkDeadLetter(_ context.Context, id uuid.UUID, _ int, _ int32, reason string) error {
	f.markDead++
	f.lastDeadID = id
	f.lastDeadReason = reason
	return nil
}

type fakeBus struct {
	publishErr error
	published  int
}

func (b *fakeBus) Publish(context.Context, string, any) error {
	if b.publishErr != nil {
		return b.publishErr
	}
	b.published++
	return nil
}

func (b *fakeBus) Subscribe(string, func(context.Context, []byte) error) error { return nil }
func (b *fakeBus) Close() error                                                { return nil }

func TestOutboxDispatcher_DispatchBatchMarksPublishedOnSuccess(t *testing.T) {
	id := uuid.New()
	repo := &fakeOutboxRepo{rows: []*entity.IngestionOutbox{{
		BaseModel:      entity.IngestionOutbox{}.BaseModel,
		EventID:        uuid.New(),
		EventType:      "document.ingestion.requested.v1",
		SchemaVersion:  "1.0.0",
		Producer:       "core-backend",
		IdempotencyKey: "idem-1",
		AccountID:      uuid.New(),
		UserID:         uuid.New(),
		Payload:        datatypes.JSONMap{"x": 1},
		ReplayCount:    0,
		AttemptCount:   0,
	}}}
	repo.rows[0].ID = id

	dispatcher := NewOutboxDispatcher(
		repo,
		&fakeBus{},
		aisvc.NewEnvelopeSigner(&core.Config{Ingestion: core.IngestionConfig{Signing: core.IngestionSigningConfig{ActiveKeyID: "k", ActiveKeySecret: "s"}}}),
		&core.Config{},
	)

	if err := dispatcher.DispatchBatch(context.Background(), 10); err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}
	if repo.markPublished != 1 {
		t.Fatalf("expected markPublished=1, got %d", repo.markPublished)
	}
	if repo.markRetry != 0 || repo.markDead != 0 {
		t.Fatalf("expected no retries/dead-letter, got retry=%d dead=%d", repo.markRetry, repo.markDead)
	}
}

func TestOutboxDispatcher_DispatchBatchSchedulesRetryOnPublishFailure(t *testing.T) {
	id := uuid.New()
	repo := &fakeOutboxRepo{rows: []*entity.IngestionOutbox{{
		EventID:        uuid.New(),
		EventType:      "document.ingestion.requested.v1",
		SchemaVersion:  "1.0.0",
		Producer:       "core-backend",
		IdempotencyKey: "idem-2",
		AccountID:      uuid.New(),
		UserID:         uuid.New(),
		Payload:        datatypes.JSONMap{"x": 1},
		ReplayCount:    1,
		AttemptCount:   1,
	}}}
	repo.rows[0].ID = id

	dispatcher := NewOutboxDispatcher(
		repo,
		&fakeBus{publishErr: errors.New("publish failed")},
		aisvc.NewEnvelopeSigner(&core.Config{Ingestion: core.IngestionConfig{Signing: core.IngestionSigningConfig{ActiveKeyID: "k", ActiveKeySecret: "s"}}}),
		&core.Config{},
	)

	if err := dispatcher.DispatchBatch(context.Background(), 10); err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}
	if repo.markRetry != 1 {
		t.Fatalf("expected markRetry=1, got %d", repo.markRetry)
	}
	if repo.lastRetryID != id {
		t.Fatalf("expected retry id %s, got %s", id, repo.lastRetryID)
	}
}

func TestOutboxDispatcher_DispatchBatchMovesToDeadLetterAfterMaxAttempts(t *testing.T) {
	id := uuid.New()
	repo := &fakeOutboxRepo{rows: []*entity.IngestionOutbox{{
		EventID:        uuid.New(),
		EventType:      "document.ingestion.requested.v1",
		SchemaVersion:  "1.0.0",
		Producer:       "core-backend",
		IdempotencyKey: "idem-3",
		AccountID:      uuid.New(),
		UserID:         uuid.New(),
		Payload:        datatypes.JSONMap{"x": 1},
		ReplayCount:    9,
		AttemptCount:   9,
	}}}
	repo.rows[0].ID = id

	dispatcher := NewOutboxDispatcher(
		repo,
		&fakeBus{publishErr: errors.New("publish failed")},
		aisvc.NewEnvelopeSigner(&core.Config{Ingestion: core.IngestionConfig{Signing: core.IngestionSigningConfig{ActiveKeyID: "k", ActiveKeySecret: "s"}}}),
		&core.Config{Ingestion: core.IngestionConfig{Dispatcher: core.IngestionDispatcherConfig{MaxAttemptsBeforeDLQ: 10}}},
	)

	if err := dispatcher.DispatchBatch(context.Background(), 10); err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}
	if repo.markDead != 1 {
		t.Fatalf("expected markDead=1, got %d", repo.markDead)
	}
	if repo.lastDeadID != id {
		t.Fatalf("expected dead-letter id %s, got %s", id, repo.lastDeadID)
	}
}
