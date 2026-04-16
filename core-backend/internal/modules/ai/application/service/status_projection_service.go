package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/entity"
	airepository "github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/repository"
	"github.com/google/uuid"
)

type StatusProjectionService struct {
	projectionRepo airepository.IngestionStatusProjectionRepository
	logger         core.Logger
}

func NewStatusProjectionService(
	projectionRepo airepository.IngestionStatusProjectionRepository,
	logger core.Logger,
) *StatusProjectionService {
	return &StatusProjectionService{
		projectionRepo: projectionRepo,
		logger:         logger,
	}
}

type StatusEventPayload struct {
	DocumentID string `json:"document_id"`
	AccountID  string `json:"account_id"`
	FromStage  string `json:"from_stage"`
	ToStage    string `json:"to_stage"`
	IsTerminal bool   `json:"is_terminal"`
	RetryCount int    `json:"retry_count"`

	ErrorMessage         *string `json:"error_message,omitempty"`
	ChunksProcessedCount *int    `json:"chunks_processed_count,omitempty"`
	ChunksFailedCount    *int    `json:"chunks_failed_count,omitempty"`
}

type StatusEventEnvelope struct {
	EventID       string             `json:"event_id"`
	EventType     string             `json:"event_type"`
	SchemaVersion string             `json:"schema_version"`
	OccurredAt    string             `json:"occurred_at"`
	Payload       StatusEventPayload `json:"payload"`
}

func (s *StatusProjectionService) ConsumeStatusEvent(ctx context.Context, envelope StatusEventEnvelope) error {
	documentID, err := uuid.Parse(envelope.Payload.DocumentID)
	if err != nil {
		return fmt.Errorf("invalid document_id: %w", err)
	}

	accountID, err := uuid.Parse(envelope.Payload.AccountID)
	if err != nil {
		return fmt.Errorf("invalid account_id: %w", err)
	}

	eventID := envelope.EventID
	if eventID == "" {
		eventID = uuid.New().String()
	}

	currentStage := s.parseStage(envelope.Payload.ToStage)
	isTerminal := currentStage == entity.IngestionStageCompleted ||
		currentStage == entity.IngestionStageFailed ||
		currentStage == entity.IngestionStageCancelled

	now := time.Now().UTC()
	var completedAt *time.Time
	if isTerminal {
		completedAt = &now
	}

	projection := &entity.IngestionStatusProjection{
		DocumentID:        documentID,
		AccountID:         accountID,
		EventID:           eventID,
		CurrentStage:      currentStage,
		IsTerminal:        isTerminal,
		StartedAt:         now,
		UpdatedAt:         now,
		CompletedAt:       completedAt,
		LastEventSequence: now.UnixMilli(),
	}

	if envelope.Payload.ErrorMessage != nil {
		projection.LastError = envelope.Payload.ErrorMessage
	}

	if envelope.Payload.ChunksProcessedCount != nil {
		projection.ChunksProcessedCount = *envelope.Payload.ChunksProcessedCount
	}

	if envelope.Payload.ChunksFailedCount != nil {
		projection.ChunksFailedCount = *envelope.Payload.ChunksFailedCount
	}

	if err := s.projectionRepo.UpsertProjection(ctx, projection); err != nil {
		s.logger.Error("Failed to upsert status projection", core.Error(err))
		return fmt.Errorf("failed to upsert projection: %w", err)
	}

	s.logger.Info("Status projection updated", core.String("document_id", documentID.String()), core.String("stage", string(currentStage)))
	return nil
}

func (s *StatusProjectionService) parseStage(stage string) entity.IngestionStage {
	switch stage {
	case "queued":
		return entity.IngestionStageQueued
	case "validating":
		return entity.IngestionStageValidating
	case "fetching":
		return entity.IngestionStageFetching
	case "chunking":
		return entity.IngestionStageChunking
	case "embedding":
		return entity.IngestionStageEmbedding
	case "indexing":
		return entity.IngestionStageIndexing
	case "completed":
		return entity.IngestionStageCompleted
	case "failed":
		return entity.IngestionStageFailed
	case "cancelled":
		return entity.IngestionStageCancelled
	default:
		return entity.IngestionStageValidating
	}
}
