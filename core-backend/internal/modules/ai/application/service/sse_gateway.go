package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	airepository "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/repository"
	"github.com/google/uuid"
)

type SSEGateway struct {
	projectionRepo airepository.IngestionStatusProjectionRepository
	logger         core.Logger
}

func NewSSEGateway(projectionRepo airepository.IngestionStatusProjectionRepository, logger core.Logger) *SSEGateway {
	return &SSEGateway{
		projectionRepo: projectionRepo,
		logger:         logger,
	}
}

type SSEDeliveryFunc func(event string, data []byte) error

func (s *SSEGateway) StreamStatusByDocument(ctx context.Context, documentID uuid.UUID, lastEventID string, sendFunc SSEDeliveryFunc) error {
	var lastSequence int64
	if lastEventID != "" {
		_, err := fmt.Sscanf(lastEventID, "%d", &lastSequence)
		if err != nil {
			lastSequence = 0
		}
	}

	pollInterval := 500 * time.Millisecond
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			projection, err := s.projectionRepo.GetByDocumentID(ctx, documentID)
			if err != nil {
				s.logger.Error("Failed to get projection for SSE", core.Error(err))
				continue
			}

			if projection == nil {
				continue
			}

			if projection.LastEventSequence <= lastSequence {
				continue
			}

			eventData, err := json.Marshal(map[string]any{
				"documentId":           projection.DocumentID,
				"accountId":            projection.AccountID,
				"userId":               projection.UserID,
				"currentStage":         string(projection.CurrentStage),
				"isTerminal":           projection.IsTerminal,
				"startedAt":            projection.StartedAt,
				"updatedAt":            projection.UpdatedAt,
				"completedAt":          projection.CompletedAt,
				"lastError":            projection.LastError,
				"chunksProcessedCount": projection.ChunksProcessedCount,
				"chunksFailedCount":    projection.ChunksFailedCount,
				"eventSequence":        projection.LastEventSequence,
			})
			if err != nil {
				s.logger.Error("Failed to marshal status event", core.Error(err))
				continue
			}

			eventID := fmt.Sprintf("%d", projection.LastEventSequence)
			if err := sendFunc(eventID, eventData); err != nil {
				return err
			}

			lastSequence = projection.LastEventSequence

			if projection.IsTerminal {
				return nil
			}
		}
	}
}

func (s *SSEGateway) StreamStatusByAccount(ctx context.Context, accountID uuid.UUID, lastEventID string, sendFunc SSEDeliveryFunc) error {
	var lastUpdatedAt time.Time
	if lastEventID != "" {
		seq, err := strconv.ParseInt(lastEventID, 10, 64)
		if err == nil {
			lastUpdatedAt = time.UnixMilli(seq)
		}
	}

	pollInterval := 3000 * time.Millisecond
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			projections, err := s.projectionRepo.GetByAccountID(ctx, accountID, 100, 0)
			if err != nil {
				s.logger.Error("Failed to list projections for SSE", core.Error(err))
				continue
			}

			maxUpdated := lastUpdatedAt
			formattedProjections := make([]map[string]any, 0, len(projections))
			for _, p := range projections {
				if !p.UpdatedAt.After(lastUpdatedAt) {
					continue
				}
				if p.UpdatedAt.After(maxUpdated) {
					maxUpdated = p.UpdatedAt
				}
				formattedProjections = append(formattedProjections, map[string]any{
					"documentId":           p.DocumentID,
					"accountId":            p.AccountID,
					"userId":               p.UserID,
					"currentStage":         string(p.CurrentStage),
					"isTerminal":           p.IsTerminal,
					"startedAt":            p.StartedAt,
					"updatedAt":            p.UpdatedAt,
					"completedAt":          p.CompletedAt,
					"lastError":            p.LastError,
					"chunksProcessedCount": p.ChunksProcessedCount,
					"chunksFailedCount":    p.ChunksFailedCount,
					"eventSequence":        p.LastEventSequence,
				})
			}

			if len(formattedProjections) == 0 {
				continue
			}

			eventData, err := json.Marshal(map[string]any{
				"projections": formattedProjections,
				"maxSequence": maxUpdated.UnixMilli(),
			})
			if err != nil {
				s.logger.Error("Failed to marshal status list event", core.Error(err))
				continue
			}

			eventID := fmt.Sprintf("%d", maxUpdated.UnixMilli())
			if err := sendFunc(eventID, eventData); err != nil {
				return err
			}

			lastUpdatedAt = maxUpdated
		}
	}
}
