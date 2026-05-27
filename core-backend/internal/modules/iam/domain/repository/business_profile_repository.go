package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type BusinessProfileRepository interface {
	sharedrepo.GenericRepository[entity.BusinessProfile]

	GetByAccountID(ctx context.Context, accountID uuid.UUID) (*entity.BusinessProfile, error)
	ExistsByAccountID(ctx context.Context, accountID uuid.UUID) (bool, error)
	GetImageURLByAccount(ctx context.Context, accountID uuid.UUID, column string) (string, error)
	UpdateImageURL(ctx context.Context, accountID uuid.UUID, column string, imageURL string) error
}
