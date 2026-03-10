package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type AccountPreferenceRepository interface {
	sharedrepo.GenericRepository[entity.AccountPreference]

	GetByAccountID(ctx context.Context, accountID uuid.UUID) (*entity.AccountPreference, error)
}
