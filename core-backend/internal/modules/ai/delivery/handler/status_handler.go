package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/repository"
	"github.com/google/uuid"
)

type StatusHandler struct {
	projectionRepo repository.IngestionStatusProjectionRepository
	documentRepo   repository.IngestionDocumentRepository
}

func NewStatusHandler(projectionRepo repository.IngestionStatusProjectionRepository, documentRepo repository.IngestionDocumentRepository) *StatusHandler {
	return &StatusHandler{projectionRepo: projectionRepo, documentRepo: documentRepo}
}

// enrichProjectionWithDocument merges document metadata from ingestion_documents into the projection response.
// Silently skips if the document lookup fails (the projection fields are still returned).
func (h *StatusHandler) enrichProjectionWithDocument(ctx context.Context, resp *dto.IngestionStatusProjectionResponse) {
	doc, err := h.documentRepo.GetByID(ctx, resp.DocumentID)
	if err != nil || doc == nil {
		return
	}
	resp.SourceFilename = doc.SourceFilename
	resp.DeclaredLanguage = doc.DeclaredLanguage
	resp.SectorIDs = doc.SectorIDs
	resp.TagIDs = doc.TagIDs
}

func mapProjectionToResponse(p *entity.IngestionStatusProjection) dto.IngestionStatusProjectionResponse {
	return dto.IngestionStatusProjectionResponse{
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

	resp := mapProjectionToResponse(projection)
	h.enrichProjectionWithDocument(ctx, &resp)

	return &dto.GetStatusByDocumentOutput{
		Body: resp,
	}, nil
}

func (h *StatusHandler) HandleListStatusByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) (*dto.ListStatusByAccountOutput, error) {
	projections, err := h.projectionRepo.GetByAccountID(ctx, accountID, limit, offset)
	if err != nil {
		return nil, err
	}

	result := make([]dto.IngestionStatusProjectionResponse, len(projections))
	for i, p := range projections {
		resp := mapProjectionToResponse(p)
		h.enrichProjectionWithDocument(ctx, &resp)
		result[i] = resp
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
		resp := mapProjectionToResponse(p)
		h.enrichProjectionWithDocument(ctx, &resp)
		result[i] = resp
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
