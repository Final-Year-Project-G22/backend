package service

import (
	"context"
	"math"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/rabbitmq"
	"github.com/google/uuid"
)

type NotificationOutboxDispatcher struct {
	outboxRepo repository.NotificationOutboxRepository
	bus        rabbitmq.Bus
	logger     core.Logger
	batchSize  int
	interval   time.Duration
	maxRetries int
}

func NewNotificationOutboxDispatcher(
	outboxRepo repository.NotificationOutboxRepository,
	bus rabbitmq.Bus,
	logger core.Logger,
) *NotificationOutboxDispatcher {
	return &NotificationOutboxDispatcher{
		outboxRepo: outboxRepo,
		bus:        bus,
		logger:     logger,
		batchSize:  50,
		interval:   5 * time.Second,
		maxRetries: 8,
	}
}

func (d *NotificationOutboxDispatcher) Start(ctx context.Context) {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := d.dispatchBatch(ctx, d.batchSize); err != nil {
				d.logger.Error("Notification outbox dispatch error", core.Error(err))
			}
		}
	}
}

func (d *NotificationOutboxDispatcher) dispatchBatch(ctx context.Context, limit int) error {
	rows, err := d.outboxRepo.ListPending(ctx, time.Now(), limit)
	if err != nil {
		return err
	}

	for _, row := range rows {
		payload := map[string]interface{}(row.Payload)

		if err := d.bus.Publish(ctx, row.EventType, payload); err != nil {
			d.handleFailure(ctx, row.ID, row.AttemptCount+1, err.Error())
			continue
		}

		_ = d.outboxRepo.MarkPublished(ctx, row.ID, time.Now())
	}

	return nil
}

func (d *NotificationOutboxDispatcher) handleFailure(ctx context.Context, id uuid.UUID, attemptCount int, reason string) {
	if attemptCount >= d.maxRetries {
		_ = d.outboxRepo.MarkDeadLetter(ctx, id, attemptCount, reason)
		d.logger.Error("Notification outbox row dead-lettered",
			core.String("id", id.String()),
			core.Int("attemptCount", attemptCount),
			core.String("reason", reason),
		)
		return
	}

	delay := d.backoffDelay(attemptCount)
	_ = d.outboxRepo.MarkRetryScheduled(ctx, id, attemptCount, time.Now().Add(delay), reason)
	d.logger.Warn("Notification outbox row scheduled for retry",
		core.String("id", id.String()),
		core.Int("attemptCount", attemptCount),
		core.String("nextAttempt", time.Now().Add(delay).String()),
		core.String("reason", reason),
	)
}

func (d *NotificationOutboxDispatcher) backoffDelay(attempt int) time.Duration {
	if attempt <= 1 {
		return 1 * time.Minute
	}

	multiplier := math.Pow(2, float64(attempt-1))
	delay := time.Duration(float64(time.Minute) * multiplier)

	if delay > 30*time.Minute {
		return 30 * time.Minute
	}
	return delay
}
