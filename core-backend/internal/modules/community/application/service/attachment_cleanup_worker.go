package service

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
)

const attachmentCleanupInterval = 15 * time.Minute
const attachmentOrphanThreshold = 1 * time.Hour

// AttachmentCleanupWorker periodically cleans up orphaned pending attachments.
// Orphaned attachments are pending attachments older than the threshold
// that were uploaded but never linked to a post.
type AttachmentCleanupWorker struct {
	attachmentUsecase usecase.AttachmentUsecase
}

// NewAttachmentCleanupWorker creates a new AttachmentCleanupWorker.
func NewAttachmentCleanupWorker(attachmentUsecase usecase.AttachmentUsecase) *AttachmentCleanupWorker {
	return &AttachmentCleanupWorker{
		attachmentUsecase: attachmentUsecase,
	}
}

// Start launches the cleanup loop in a goroutine.
// Cancellation is handled via ctx.Done().
func (w *AttachmentCleanupWorker) Start(ctx context.Context) {
	go w.runCleanupLoop(ctx)
}

// runCleanupLoop runs the cleanup periodically every 15 minutes.
// Errors are logged but do not interrupt the loop.
func (w *AttachmentCleanupWorker) runCleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(attachmentCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			threshold := time.Now().UTC().Add(-attachmentOrphanThreshold)
			_ = w.attachmentUsecase.CleanupPending(ctx, threshold)
		}
	}
}
