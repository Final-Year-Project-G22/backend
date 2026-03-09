package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type CommunityPreferenceRepository interface {
	sharedrepo.GenericRepository[entity.CommunityPreference]

	GetByAccountID(ctx context.Context, accountID uuid.UUID) (*entity.CommunityPreference, error)
}
