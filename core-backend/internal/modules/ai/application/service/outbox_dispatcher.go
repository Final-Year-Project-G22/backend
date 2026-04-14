package service

import (
	"context"
	"time"

	aievent "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/event"
	airepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/repository"
	aisvc "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/service"
	"github.com/Final-Year-Project-G22/backend/core/pkg/rabbitmq"
)

type OutboxDispatcher struct {
	outboxRepo airepo.IngestionOutboxRepository
	bus        rabbitmq.Bus
	signer     aisvc.EnvelopeSigner
}

func NewOutboxDispatcher(
	outboxRepo airepo.IngestionOutboxRepository,
	bus rabbitmq.Bus,
	signer aisvc.EnvelopeSigner,
) *OutboxDispatcher {
	return &OutboxDispatcher{outboxRepo: outboxRepo, bus: bus, signer: signer}
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
			_ = d.outboxRepo.MarkFailed(ctx, row.ID, row.AttemptCount+1, time.Now().Add(30*time.Second), row.ReplayCount+1, signErr.Error())
			continue
		}

		envelope["key_id"] = keyID
		envelope["signature"] = string(signature)

		if publishErr := d.bus.Publish(ctx, aievent.DocumentIngestionRequestedV1, envelope); publishErr != nil {
			_ = d.outboxRepo.MarkFailed(ctx, row.ID, row.AttemptCount+1, time.Now().Add(30*time.Second), row.ReplayCount+1, publishErr.Error())
			continue
		}

		_ = d.outboxRepo.MarkPublished(ctx, row.ID, time.Now(), signature)
	}

	return nil
}
