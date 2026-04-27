package service

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
)

const (
	deliveryPollInterval = 5 * time.Second
	deliveryBatchSize    = 50
	expiryInterval       = 1 * time.Hour
)

// DeliveryWorker runs background goroutines for the notification delivery pipeline.
// It polls the notification queue for pending items and periodically cleans up
// expired inbox entries.
type DeliveryWorker struct {
	deliveryUC usecase.NotificationDeliveryUsecase
	inboxUC    usecase.NotificationInboxUsecase
}

// NewDeliveryWorker creates a new DeliveryWorker.
func NewDeliveryWorker(
	deliveryUC usecase.NotificationDeliveryUsecase,
	inboxUC usecase.NotificationInboxUsecase,
) *DeliveryWorker {
	return &DeliveryWorker{
		deliveryUC: deliveryUC,
		inboxUC:    inboxUC,
	}
}

// Start launches the delivery poll and inbox expiry goroutines.
// It returns immediately; cancellation is handled via ctx.Done().
func (w *DeliveryWorker) Start(ctx context.Context) {
	go w.runDeliveryLoop(ctx)
	go w.runExpiryLoop(ctx)
}

// runDeliveryLoop polls ProcessQueue every 5 seconds for pending notifications.
// Errors are swallowed to avoid interrupting the loop.
func (w *DeliveryWorker) runDeliveryLoop(ctx context.Context) {
	ticker := time.NewTicker(deliveryPollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = w.deliveryUC.ProcessQueue(ctx, deliveryBatchSize)
		}
	}
}

// runExpiryLoop calls ExpireOld every hour to clean up expired inbox entries.
// Errors are swallowed to avoid interrupting the loop.
func (w *DeliveryWorker) runExpiryLoop(ctx context.Context) {
	ticker := time.NewTicker(expiryInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = w.inboxUC.ExpireOld(ctx, time.Now().UTC())
		}
	}
}
