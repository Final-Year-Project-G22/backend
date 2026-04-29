package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type AccountRepository interface {
	sharedrepo.GenericRepository[entity.Account]

	GetByEmailNormalized(ctx context.Context, email string) (*entity.Account, error)
	GetByUsernameNormalized(ctx context.Context, username string) (*entity.Account, error)
	GetByEmailOrUsername(ctx context.Context, identifier string) (*entity.Account, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Account, error)
	ExistsByEmailNormalized(ctx context.Context, email string) (bool, error)
	ExistsByUsernameNormalized(ctx context.Context, username string) (bool, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.AccountStatus) error
	MarkEmailVerifiedAndActivate(ctx context.Context, id uuid.UUID) error

	// FindBySegment returns accounts matching segment filters.
	// Supported keys: "status", "created_after" (RFC3339), "created_before" (RFC3339)
	FindBySegment(ctx context.Context, segment map[string]interface{}) ([]*entity.Account, error)
}
