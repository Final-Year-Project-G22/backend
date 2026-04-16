package service

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	airepository "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/repository"
	"github.com/google/uuid"
)

const (
	DefaultSnapshotInterval = 30 * time.Second
	DefaultProgressDelta    = 10
)

type PeriodicSnapshotEmitter struct {
	projectionRepo   airepository.IngestionStatusProjectionRepository
	logger           core.Logger
	snapshotInterval time.Duration
	progressDelta    int
}

func NewPeriodicSnapshotEmitter(
	projectionRepo airepository.IngestionStatusProjectionRepository,
	logger core.Logger,
) *PeriodicSnapshotEmitter {
	return &PeriodicSnapshotEmitter{
		projectionRepo:   projectionRepo,
		logger:           logger,
		snapshotInterval: DefaultSnapshotInterval,
		progressDelta:    DefaultProgressDelta,
	}
}

func (e *PeriodicSnapshotEmitter) EmitSnapshotForDocument(ctx context.Context, documentID uuid.UUID) error {
	projection, err := e.projectionRepo.GetByDocumentID(ctx, documentID)
	if err != nil {
		return err
	}

	if projection == nil {
		return nil
	}

	if projection.IsTerminal {
		return nil
	}

	e.logger.Debug("Periodic snapshot emitted",
		core.String("document_id", documentID.String()),
		core.String("stage", string(projection.CurrentStage)),
		core.Int("chunks_processed", projection.ChunksProcessedCount))

	return nil
}

func (e *PeriodicSnapshotEmitter) EmitSnapshotForAccount(ctx context.Context, accountID uuid.UUID) error {
	projections, err := e.projectionRepo.GetByAccountID(ctx, accountID, 100, 0)
	if err != nil {
		return err
	}

	if len(projections) == 0 {
		return nil
	}

	var activeCount int
	for _, p := range projections {
		if !p.IsTerminal {
			activeCount++
		}
	}

	e.logger.Debug("Periodic snapshot emitted for account",
		core.String("account_id", accountID.String()),
		core.Int("active_projections", activeCount))

	return nil
}

func (e *PeriodicSnapshotEmitter) StartPeriodicEmitter(ctx context.Context, accountID uuid.UUID) {
	ticker := time.NewTicker(e.snapshotInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = e.EmitSnapshotForAccount(ctx, accountID)
		}
	}
}

func (e *PeriodicSnapshotEmitter) SnapshotInterval() time.Duration {
	return e.snapshotInterval
}

func (e *PeriodicSnapshotEmitter) ProgressDelta() int {
	return e.progressDelta
}

func (e *PeriodicSnapshotEmitter) ShouldEmit(progressBefore, progressAfter int) bool {
	return (progressAfter - progressBefore) >= e.progressDelta
}
