package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/entity"
	"github.com/google/uuid"
)

type IngestionStatusProjectionRepository interface {
	UpsertProjection(ctx context.Context, projection *entity.IngestionStatusProjection) error
	GetByDocumentID(ctx context.Context, documentID uuid.UUID) (*entity.IngestionStatusProjection, error)
	GetByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*entity.IngestionStatusProjection, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.IngestionStatusProjection, error)
}
