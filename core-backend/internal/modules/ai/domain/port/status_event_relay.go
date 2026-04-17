package port

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/entity"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type StatusEventRelay interface {
	PublishStatusUpdate(ctx context.Context, projection *entity.IngestionStatusProjection) error
	PublishAccountUpdate(ctx context.Context, accountID uuid.UUID, projections []*entity.IngestionStatusProjection) error
	SubscribeToDocument(ctx context.Context, documentID uuid.UUID) *redis.PubSub
	GetReplayEvents(ctx context.Context, documentID uuid.UUID, sinceSequence int64) ([]map[string]any, error)
}
