package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type CategoryRepository interface {
	sharedrepo.GenericRepository[entity.GuideCategory]

	GetBySlug(ctx context.Context, parentID *uuid.UUID, slug string) (*entity.GuideCategory, error)
	ListTree(ctx context.Context, includeInactive bool) ([]*entity.GuideCategory, error)

	GetConditions(ctx context.Context, categoryID uuid.UUID) ([]*entity.GuideCategoryCondition, error)
	AddCondition(ctx context.Context, cond *entity.GuideCategoryCondition) error
	RemoveCondition(ctx context.Context, condID uuid.UUID) error

	GetTranslations(ctx context.Context, categoryID uuid.UUID) ([]*entity.GuideCategoryTranslation, error)
	UpsertTranslation(ctx context.Context, trans *entity.GuideCategoryTranslation) error
	DeleteTranslation(ctx context.Context, categoryID uuid.UUID, language string) error
}
