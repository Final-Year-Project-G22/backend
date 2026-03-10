package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type PermissionRepository interface {
	sharedrepo.GenericRepository[entity.Permission]

	GetByCode(ctx context.Context, code string) (*entity.Permission, error)
	ListByCodes(ctx context.Context, codes []string) ([]*entity.Permission, error)
	ListByRoleID(ctx context.Context, roleID uuid.UUID) ([]*entity.Permission, error)
}
