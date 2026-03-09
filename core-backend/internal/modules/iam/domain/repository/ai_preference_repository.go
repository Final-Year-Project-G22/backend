package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type AIPreferenceRepository interface {
	sharedrepo.GenericRepository[entity.AIPreference]

	GetByAccountID(ctx context.Context, accountID uuid.UUID) (*entity.AIPreference, error)
}
