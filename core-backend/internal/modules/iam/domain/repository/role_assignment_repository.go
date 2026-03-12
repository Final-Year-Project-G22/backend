package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type RoleAssignmentRepository interface {
	sharedrepo.GenericRepository[entity.RoleAssignment]

	ListByAccountID(ctx context.Context, accountID uuid.UUID) ([]*entity.RoleAssignment, error)
	ExistsByAccountAndRole(ctx context.Context, accountID, roleID uuid.UUID) (bool, error)
	GetByAccountAndRole(ctx context.Context, accountID, roleID uuid.UUID) (*entity.RoleAssignment, error)
	Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time, reason *string) error
}
