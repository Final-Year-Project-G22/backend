package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
	"github.com/google/uuid"
)

type TierService interface {
	HasAccess(ctx context.Context, accountID uuid.UUID, requiredTier entity.TierAccess) (bool, error)
}
