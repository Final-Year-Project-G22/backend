package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type OAuthIdentityRepository interface {
	sharedrepo.GenericRepository[entity.OAuthIdentity]

	GetByProviderSubject(ctx context.Context, provider, subject string) (*entity.OAuthIdentity, error)
	ListByAccountID(ctx context.Context, accountID uuid.UUID) ([]*entity.OAuthIdentity, error)
}
