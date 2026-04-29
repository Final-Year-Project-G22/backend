package repository

import (
	"context"

	"github.com/google/uuid"
)

type AccountReader interface {
	FindAll(ctx context.Context) ([]uuid.UUID, error)
	FindBySegment(ctx context.Context, segment map[string]interface{}) ([]uuid.UUID, error)
}
