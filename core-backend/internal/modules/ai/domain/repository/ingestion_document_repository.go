package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/entity"
	"github.com/google/uuid"
)

type IngestionDocumentRepository interface {
	Create(ctx context.Context, doc *entity.IngestionDocument) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.IngestionDocument, error)
	GetByIdempotencyKey(ctx context.Context, accountID uuid.UUID, idempotencyKey string) (*entity.IngestionDocument, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.IngestionDocumentStatus, lastError *string) error
	SoftDelete(ctx context.Context, id uuid.UUID, accountID uuid.UUID) error
}
