package port

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	"github.com/google/uuid"
)

type AskRequest struct {
	RequestID uuid.UUID
	UserID    uuid.UUID
	AccountID uuid.UUID
	Query     string
	Language  constants.Locale
	SessionID *uuid.UUID
	TopK      int
}

type Citation struct {
	DocumentID uuid.UUID
	ChunkID    uuid.UUID
	SourceType string
	Title      *string
	Score      float64
}

type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type AskResponse struct {
	RequestID uuid.UUID
	SessionID uuid.UUID
	Answer    string
	Citations []Citation
	Usage     Usage
	Model     string
	LatencyMS int
}

type AIInferencePort interface {
	Ask(ctx context.Context, req AskRequest) (AskResponse, error)
}
