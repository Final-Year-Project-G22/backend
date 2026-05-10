package tier

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/usecase"
	"github.com/google/uuid"
)

type tierServiceStub struct{}

func NewTierServiceStub() usecase.TierService {
	return &tierServiceStub{}
}

func (s *tierServiceStub) HasAccess(ctx context.Context, accountID uuid.UUID, requiredTier entity.TierAccess) (bool, error) {
	return true, nil
}
