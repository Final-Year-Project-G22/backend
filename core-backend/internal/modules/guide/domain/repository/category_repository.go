package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type CategoryRepository interface {
	sharedrepo.GenericRepository[entity.GuideCategory]

	GetBySlug(ctx context.Context, parentID *uuid.UUID, slug string, locale constants.Locale) (*entity.GuideCategory, error)
	ListTree(ctx context.Context, includeInactive bool, locale constants.Locale) ([]*entity.GuideCategory, error)

	GetConditions(ctx context.Context, categoryID uuid.UUID) ([]*entity.GuideCategoryCondition, error)
	AddCondition(ctx context.Context, cond *entity.GuideCategoryCondition) error
	RemoveCondition(ctx context.Context, condID uuid.UUID) error

	GetTranslations(ctx context.Context, categoryID uuid.UUID) ([]*entity.GuideCategoryTranslation, error)
	UpsertTranslation(ctx context.Context, trans *entity.GuideCategoryTranslation) error
	DeleteTranslation(ctx context.Context, categoryID uuid.UUID, language string) error
}
