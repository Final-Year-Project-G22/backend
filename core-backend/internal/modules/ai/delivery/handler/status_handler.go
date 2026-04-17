package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/repository"
	"github.com/google/uuid"
)

type StatusHandler struct {
	projectionRepo repository.IngestionStatusProjectionRepository
}

func NewStatusHandler(projectionRepo repository.IngestionStatusProjectionRepository) *StatusHandler {
	return &StatusHandler{projectionRepo: projectionRepo}
}

func (h *StatusHandler) HandleGetStatusByDocumentID(ctx context.Context, documentID uuid.UUID) (*dto.GetStatusByDocumentOutput, error) {
	projection, err := h.projectionRepo.GetByDocumentID(ctx, documentID)
	if err != nil {
		return nil, err
	}

	if projection == nil {
		return &dto.GetStatusByDocumentOutput{
			Body: dto.IngestionStatusProjectionResponse{},
		}, nil
	}

	return &dto.GetStatusByDocumentOutput{
		Body: dto.IngestionStatusProjectionResponse{
			DocumentID:           projection.DocumentID,
			AccountID:            projection.AccountID,
			UserID:               projection.UserID,
			CurrentStage:         string(projection.CurrentStage),
			IsTerminal:           projection.IsTerminal,
			StartedAt:            projection.StartedAt,
			UpdatedAt:            projection.UpdatedAt,
			CompletedAt:          projection.CompletedAt,
			LastError:            projection.LastError,
			ChunksProcessedCount: projection.ChunksProcessedCount,
			ChunksFailedCount:    projection.ChunksFailedCount,
			EventSequence:        projection.LastEventSequence,
		},
	}, nil
}

func (h *StatusHandler) HandleListStatusByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) (*dto.ListStatusByAccountOutput, error) {
	projections, err := h.projectionRepo.GetByAccountID(ctx, accountID, limit, offset)
	if err != nil {
		return nil, err
	}

	result := make([]dto.IngestionStatusProjectionResponse, len(projections))
	for i, p := range projections {
		result[i] = dto.IngestionStatusProjectionResponse{
			DocumentID:           p.DocumentID,
			AccountID:            p.AccountID,
			UserID:               p.UserID,
			CurrentStage:         string(p.CurrentStage),
			IsTerminal:           p.IsTerminal,
			StartedAt:            p.StartedAt,
			UpdatedAt:            p.UpdatedAt,
			CompletedAt:          p.CompletedAt,
			LastError:            p.LastError,
			ChunksProcessedCount: p.ChunksProcessedCount,
			ChunksFailedCount:    p.ChunksFailedCount,
			EventSequence:        p.LastEventSequence,
		}
	}

	return &dto.ListStatusByAccountOutput{
		Body: struct {
			Projections []dto.IngestionStatusProjectionResponse `json:"projections" doc:"List of status projections"`
			Total       int                                     `json:"total" doc:"Total count for pagination"`
		}{
			Projections: result,
			Total:       len(result),
		},
	}, nil
}

func (h *StatusHandler) HandleListStatusByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) (*dto.ListStatusByUserOutput, error) {
	projections, err := h.projectionRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	result := make([]dto.IngestionStatusProjectionResponse, len(projections))
	for i, p := range projections {
		result[i] = dto.IngestionStatusProjectionResponse{
			DocumentID:           p.DocumentID,
			AccountID:            p.AccountID,
			UserID:               p.UserID,
			CurrentStage:         string(p.CurrentStage),
			IsTerminal:           p.IsTerminal,
			StartedAt:            p.StartedAt,
			UpdatedAt:            p.UpdatedAt,
			CompletedAt:          p.CompletedAt,
			LastError:            p.LastError,
			ChunksProcessedCount: p.ChunksProcessedCount,
			ChunksFailedCount:    p.ChunksFailedCount,
			EventSequence:        p.LastEventSequence,
		}
	}

	return &dto.ListStatusByUserOutput{
		Body: struct {
			Projections []dto.IngestionStatusProjectionResponse `json:"projections" doc:"List of status projections"`
			Total       int                                     `json:"total" doc:"Total count for pagination"`
		}{
			Projections: result,
			Total:       len(result),
		},
	}, nil
}
