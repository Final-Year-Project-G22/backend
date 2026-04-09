package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type CommunityCategoryUsecase interface {
	CreateCategory(ctx context.Context, actorID uuid.UUID, input CreateCategoryInput) (*entity.CommunityCategory, error)
	UpdateCategory(ctx context.Context, actorID, id uuid.UUID, input UpdateCategoryInput) (*entity.CommunityCategory, error)
	DeleteCategory(ctx context.Context, actorID, id uuid.UUID) error
	ListCategories(ctx context.Context, includeInactive bool, q query.QueryOptions) ([]*entity.CommunityCategory, error)
	GetCategory(ctx context.Context, id uuid.UUID) (*entity.CommunityCategory, error)
}
