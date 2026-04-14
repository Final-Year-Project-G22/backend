package service

import (
	"context"
	"math"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	aievent "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/event"
	airepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/repository"
	aisvc "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/service"
	"github.com/Final-Year-Project-G22/backend/core/pkg/rabbitmq"
	"github.com/google/uuid"
)

type OutboxDispatcher struct {
	outboxRepo airepo.IngestionOutboxRepository
	bus        rabbitmq.Bus
	signer     aisvc.EnvelopeSigner
	cfg        core.IngestionDispatcherConfig
}

func NewOutboxDispatcher(
	outboxRepo airepo.IngestionOutboxRepository,
	bus rabbitmq.Bus,
	signer aisvc.EnvelopeSigner,
	cfg *core.Config,
) *OutboxDispatcher {
	defaultCfg := core.IngestionDispatcherConfig{
		BatchSize:            50,
		Interval:             5 * time.Second,
		RetryBaseDelay:       30 * time.Second,
		RetryMaxDelay:        10 * time.Minute,
		MaxAttemptsBeforeDLQ: 10,
	}

	if cfg != nil {
		if cfg.Ingestion.Dispatcher.BatchSize > 0 {
			defaultCfg.BatchSize = cfg.Ingestion.Dispatcher.BatchSize
		}
		if cfg.Ingestion.Dispatcher.Interval > 0 {
			defaultCfg.Interval = cfg.Ingestion.Dispatcher.Interval
		}
		if cfg.Ingestion.Dispatcher.RetryBaseDelay > 0 {
			defaultCfg.RetryBaseDelay = cfg.Ingestion.Dispatcher.RetryBaseDelay
		}
		if cfg.Ingestion.Dispatcher.RetryMaxDelay > 0 {
			defaultCfg.RetryMaxDelay = cfg.Ingestion.Dispatcher.RetryMaxDelay
		}
		if cfg.Ingestion.Dispatcher.MaxAttemptsBeforeDLQ > 0 {
			defaultCfg.MaxAttemptsBeforeDLQ = cfg.Ingestion.Dispatcher.MaxAttemptsBeforeDLQ
		}
	}

	return &OutboxDispatcher{outboxRepo: outboxRepo, bus: bus, signer: signer, cfg: defaultCfg}
}

func (d *OutboxDispatcher) BatchSize() int {
	return d.cfg.BatchSize
}

func (d *OutboxDispatcher) Interval() time.Duration {
	return d.cfg.Interval
}

func (d *OutboxDispatcher) DispatchBatch(ctx context.Context, limit int) error {
	rows, err := d.outboxRepo.ListPending(ctx, time.Now(), limit)
	if err != nil {
		return err
	}

	for _, row := range rows {
		envelope := map[string]any{
			"event_id":        row.EventID.String(),
			"event_type":      row.EventType,
			"schema_version":  row.SchemaVersion,
			"occurred_at":     row.CreatedAt,
			"producer":        row.Producer,
			"idempotency_key": row.IdempotencyKey,
			"account_id":      row.AccountID.String(),
			"user_id":         row.UserID.String(),
			"replay_count":    row.ReplayCount,
			"payload":         row.Payload,
		}

		if row.BatchID != nil {
			envelope["batch_id"] = row.BatchID.String()
		}

		signature, keyID, signErr := d.signer.SignEnvelope(ctx, envelope)
		if signErr != nil {
			d.handleFailure(ctx, row.ID, row.AttemptCount+1, row.ReplayCount+1, signErr.Error())
			continue
		}

		envelope["key_id"] = keyID
		envelope["signature"] = string(signature)

		if publishErr := d.bus.Publish(ctx, aievent.DocumentIngestionRequestedV1, envelope); publishErr != nil {
			d.handleFailure(ctx, row.ID, row.AttemptCount+1, row.ReplayCount+1, publishErr.Error())
			continue
		}

		_ = d.outboxRepo.MarkPublished(ctx, row.ID, time.Now(), signature)
	}

	return nil
}

func (d *OutboxDispatcher) handleFailure(ctx context.Context, id uuid.UUID, attemptCount int, replayCount int32, reason string) {
	if attemptCount >= d.cfg.MaxAttemptsBeforeDLQ {
		_ = d.outboxRepo.MarkDeadLetter(ctx, id, attemptCount, replayCount, reason)
		return
	}

	delay := d.backoffDelay(attemptCount)
	_ = d.outboxRepo.MarkRetryScheduled(ctx, id, attemptCount, time.Now().Add(delay), replayCount, reason)
}

func (d *OutboxDispatcher) backoffDelay(attempt int) time.Duration {
	if attempt <= 1 {
		return d.cfg.RetryBaseDelay
	}

	multiplier := math.Pow(2, float64(attempt-1))
	delay := time.Duration(float64(d.cfg.RetryBaseDelay) * multiplier)
	if delay > d.cfg.RetryMaxDelay {
		return d.cfg.RetryMaxDelay
	}
	return delay
}
