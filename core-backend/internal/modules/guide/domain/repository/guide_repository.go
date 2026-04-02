package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type GuideRepository interface {
	sharedrepo.GenericRepository[entity.Guide]

	GetBySlug(ctx context.Context, categoryID uuid.UUID, slug string, locale constants.Locale) (*entity.Guide, error)
	ListByCategory(ctx context.Context, categoryID uuid.UUID, q query.QueryOptions, locale constants.Locale) ([]*entity.Guide, error)
	Search(ctx context.Context, keyword string, q query.QueryOptions, locale constants.Locale) ([]*entity.Guide, error)

	GetConditions(ctx context.Context, guideID uuid.UUID) ([]*entity.GuideCondition, error)
	AddCondition(ctx context.Context, cond *entity.GuideCondition) error
	RemoveCondition(ctx context.Context, condID uuid.UUID) error

	GetTranslations(ctx context.Context, guideID uuid.UUID) ([]*entity.GuideTranslation, error)
	UpsertTranslation(ctx context.Context, trans *entity.GuideTranslation) error
	DeleteTranslation(ctx context.Context, guideID uuid.UUID, language string) error
}
