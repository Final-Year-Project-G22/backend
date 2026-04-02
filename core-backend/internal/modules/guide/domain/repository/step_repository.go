package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type StepRepository interface {
	sharedrepo.GenericRepository[entity.GuideStep]

	GetBySlug(ctx context.Context, guideID uuid.UUID, slug string, locale constants.Locale) (*entity.GuideStep, error)
	ListByGuide(ctx context.Context, guideID uuid.UUID, q query.QueryOptions, locale constants.Locale) ([]*entity.GuideStep, error)
	Reorder(ctx context.Context, guideID uuid.UUID, stepIDsInOrder []uuid.UUID) error

	GetConditions(ctx context.Context, stepID uuid.UUID) ([]*entity.StepCondition, error)
	AddCondition(ctx context.Context, cond *entity.StepCondition) error
	RemoveCondition(ctx context.Context, condID uuid.UUID) error

	GetDependencies(ctx context.Context, stepID uuid.UUID) ([]*entity.StepDependency, error)
	AddDependency(ctx context.Context, dep *entity.StepDependency) error
	RemoveDependency(ctx context.Context, depID uuid.UUID) error

	GetTranslations(ctx context.Context, stepID uuid.UUID) ([]*entity.GuideStepTranslation, error)
	UpsertTranslation(ctx context.Context, trans *entity.GuideStepTranslation) error
	DeleteTranslation(ctx context.Context, stepID uuid.UUID, language string) error

	GetVersions(ctx context.Context, stepID uuid.UUID, q query.QueryOptions) ([]*entity.GuideStepVersion, error)
	GetVersion(ctx context.Context, stepID uuid.UUID, version int) (*entity.GuideStepVersion, error)
	RestoreVersion(ctx context.Context, stepID uuid.UUID, version int) error
}
