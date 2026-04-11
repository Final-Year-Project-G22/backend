package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type CommunityCategoryRepository interface {
	sharedrepo.GenericRepository[entity.CommunityCategory]

	GetBySlug(ctx context.Context, parentID *uuid.UUID, slug string) (*entity.CommunityCategory, error)
	ListTree(ctx context.Context, includeInactive bool) ([]*entity.CommunityCategory, error)
	ListActive(ctx context.Context, q query.QueryOptions) ([]*entity.CommunityCategory, error)
	ExistsBySlug(ctx context.Context, parentID *uuid.UUID, slug string) (bool, error)
}
