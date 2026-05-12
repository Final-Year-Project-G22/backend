package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type GuideManagementUseCase interface {
	ListGuides(ctx context.Context, q query.QueryOptions, locale constants.Locale) (sharedrepo.PaginatedResult[entity.Guide], error)
	GetGuideDetail(ctx context.Context, guideID uuid.UUID, locale constants.Locale) (*entity.Guide, error)
	ListGuideSteps(ctx context.Context, guideID uuid.UUID, q query.QueryOptions, locale constants.Locale) (sharedrepo.PaginatedResult[entity.GuideStep], error)

	CreateGuide(ctx context.Context, input CreateGuideInput) (*entity.Guide, error)
	UpdateGuide(ctx context.Context, id uuid.UUID, input UpdateGuideInput) error
	DeleteGuide(ctx context.Context, id uuid.UUID) error
	AddGuideCondition(ctx context.Context, guideID uuid.UUID, cond ConditionInput) error
	RemoveGuideCondition(ctx context.Context, condID uuid.UUID) error
	SetGuideTranslations(ctx context.Context, guideID uuid.UUID, translations []TranslationInput, merge bool) error

	CreateStep(ctx context.Context, input CreateStepInput) (*entity.GuideStep, error)
	UpdateStep(ctx context.Context, id uuid.UUID, input UpdateStepInput) error
	DeleteStep(ctx context.Context, id uuid.UUID) error
	ReorderSteps(ctx context.Context, guideID uuid.UUID, stepIDsInOrder []uuid.UUID) error
	AddStepCondition(ctx context.Context, stepID uuid.UUID, cond ConditionInput) error
	RemoveStepCondition(ctx context.Context, condID uuid.UUID) error
	AddStepDependency(ctx context.Context, stepID, requiredStepID uuid.UUID, depType entity.DependencyType) error
	RemoveStepDependency(ctx context.Context, depID uuid.UUID) error
	SetStepTranslations(ctx context.Context, stepID uuid.UUID, translations []StepTranslationInput, merge bool) error

	GetStepVersions(ctx context.Context, stepID uuid.UUID, q query.QueryOptions) ([]*entity.GuideStepVersion, error)
	RevertStepToVersion(ctx context.Context, stepID uuid.UUID, version int) error
}

type JourneyManagementUseCase interface {
	InvalidateUserJourney(ctx context.Context, accountID, userID, guideID uuid.UUID) error
	InvalidateAllJourneysForGuide(ctx context.Context, guideID uuid.UUID) error
}
