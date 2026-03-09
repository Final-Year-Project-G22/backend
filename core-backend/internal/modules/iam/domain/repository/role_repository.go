package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
)

type RoleRepository interface {
	sharedrepo.GenericRepository[entity.Role]

	GetByCode(ctx context.Context, code string) (*entity.Role, error)
	ListByCodes(ctx context.Context, codes []string) ([]*entity.Role, error)
}
