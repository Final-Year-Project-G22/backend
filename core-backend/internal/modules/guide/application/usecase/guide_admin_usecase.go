package usecase

import (
	"context"
	"errors"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/entity"
	guideerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	sharedRepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type guideAdminUsecase struct {
	catRepo      repository.CategoryRepository
	guideRepo    repository.GuideRepository
	stepRepo     repository.StepRepository
	progressRepo repository.ProgressRepository
	transactor   sharedRepo.Transactor
	logger       core.Logger
}

func NewGuideAdminUsecase(
	catRepo repository.CategoryRepository,
	guideRepo repository.GuideRepository,
	stepRepo repository.StepRepository,
	progressRepo repository.ProgressRepository,
	transactor sharedRepo.Transactor,
	logger core.Logger,
) usecase.GuideManagementUseCase {
	return &guideAdminUsecase{
		catRepo:      catRepo,
		guideRepo:    guideRepo,
		stepRepo:     stepRepo,
		progressRepo: progressRepo,
		transactor:   transactor,
		logger:       logger,
	}
}

func (u *guideAdminUsecase) ListCategoriesTree(ctx context.Context, includeInactive bool, locale constants.Locale) ([]*entity.GuideCategory, error) {
	return u.catRepo.ListTree(ctx, includeInactive, locale)
}

func (u *guideAdminUsecase) ListGuides(ctx context.Context, categoryID *uuid.UUID, q query.QueryOptions) (sharedRepo.PaginatedResult[entity.Guide], error) {
	if categoryID != nil {
		if q.Filters == nil {
			q.Filters = make(map[string]interface{})
		}
		q.Filters["category_id"] = *categoryID
	}
	if q.Preload == nil {
		q.Preload = []string{"Translations"}
	}
	return u.guideRepo.FindAll(ctx, q), nil
}

func (u *guideAdminUsecase) ListGuideSteps(ctx context.Context, guideID uuid.UUID, q query.QueryOptions) (sharedRepo.PaginatedResult[entity.GuideStep], error) {
	if q.Filters == nil {
		q.Filters = make(map[string]interface{})
	}
	q.Filters["guide_id"] = guideID
	if q.Preload == nil {
		q.Preload = []string{"Translations"}
	}
	return u.stepRepo.FindAll(ctx, q), nil
}

func (u *guideAdminUsecase) GetGuideDetail(ctx context.Context, guideID uuid.UUID, locale constants.Locale) (*entity.Guide, error) {
	guide, err := u.guideRepo.GetByID(ctx, guideID)
	if err != nil {
		return nil, err
	}

	if locale != "" {
		translations, err := u.guideRepo.GetTranslations(ctx, guideID)
		if err != nil {
			return nil, err
		}
		guide.Translations = []entity.GuideTranslation{}
		for _, t := range translations {
			if t.Language == string(locale) {
				guide.Translations = append(guide.Translations, *t)
			}
		}
		if len(guide.Translations) == 0 {
			for _, t := range translations {
				if t.Language == string(constants.LocaleEnglish) {
					guide.Translations = append(guide.Translations, *t)
				}
			}
		}
	}

	conditions, err := u.guideRepo.GetConditions(ctx, guideID)
	if err != nil {
		return nil, err
	}
	guide.Conditions = []entity.GuideCondition{}
	for _, c := range conditions {
		guide.Conditions = append(guide.Conditions, *c)
	}
	return guide, nil
}

func (u *guideAdminUsecase) CreateCategory(ctx context.Context, input usecase.CreateCategoryInput) (*entity.GuideCategory, error) {
	var category *entity.GuideCategory
	err := u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		category = &entity.GuideCategory{
			Slug:             input.Slug,
			Icon:             input.Icon,
			SortOrder:        input.SortOrder,
			ParentCategoryID: input.ParentID,
		}
		if err := u.catRepo.Create(txCtx, category); err != nil {
			return err
		}

		if err := u.setCategoryTranslations(txCtx, category.ID, input.Translations); err != nil {
			return err
		}
		return u.setCategoryConditions(txCtx, category.ID, input.Conditions)
	})
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (u *guideAdminUsecase) UpdateCategory(ctx context.Context, id uuid.UUID, input usecase.UpdateCategoryInput) error {
	return u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		category, err := u.catRepo.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		if input.Slug != nil {
			category.Slug = *input.Slug
		}
		if input.Icon != nil {
			category.Icon = input.Icon
		}
		if input.SortOrder != nil {
			category.SortOrder = *input.SortOrder
		}
		if input.ParentID != nil {
			category.ParentCategoryID = input.ParentID
		}
		if err := u.catRepo.Update(txCtx, category); err != nil {
			return err
		}
		if input.Translations != nil {
			if err := u.setCategoryTranslations(txCtx, id, input.Translations); err != nil {
				return err
			}
		}
		if input.Conditions != nil {
			return u.setCategoryConditions(txCtx, id, input.Conditions)
		}
		return nil
	})
}

func (u *guideAdminUsecase) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return u.catRepo.Delete(ctx, id)
}

func (u *guideAdminUsecase) AddCategoryCondition(ctx context.Context, categoryID uuid.UUID, cond usecase.ConditionInput) error {
	condition := &entity.GuideCategoryCondition{
		CategoryID:     categoryID,
		ConditionType:  cond.ConditionType,
		Operator:       cond.Operator,
		ConditionValue: toJSONMap(cond.ConditionValue),
		IsInverse:      cond.IsInverse,
	}
	return u.catRepo.AddCondition(ctx, condition)
}

func (u *guideAdminUsecase) RemoveCategoryCondition(ctx context.Context, condID uuid.UUID) error {
	return u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := u.catRepo.RemoveCondition(txCtx, condID); err != nil {
			return err
		}
		return nil
	})
}

func (u *guideAdminUsecase) SetCategoryTranslations(ctx context.Context, categoryID uuid.UUID, translations []usecase.TranslationInput) error {
	return u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		return u.setCategoryTranslations(txCtx, categoryID, translations)
	})
}

func (u *guideAdminUsecase) CreateGuide(ctx context.Context, input usecase.CreateGuideInput) (*entity.Guide, error) {
	var guide *entity.Guide
	err := u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		guide = &entity.Guide{
			CategoryID: input.CategoryID,
			Slug:       input.Slug,
			Icon:       input.Icon,
			SortOrder:  input.SortOrder,
		}
		if err := u.guideRepo.Create(txCtx, guide); err != nil {
			return err
		}
		if err := u.setGuideTranslations(txCtx, guide.ID, input.Translations); err != nil {
			return err
		}
		return u.setGuideConditions(txCtx, guide.ID, input.Conditions)
	})
	if err != nil {
		return nil, err
	}
	return guide, nil
}

func (u *guideAdminUsecase) UpdateGuide(ctx context.Context, id uuid.UUID, input usecase.UpdateGuideInput) error {
	return u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		guide, err := u.guideRepo.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		if input.CategoryID != nil {
			guide.CategoryID = *input.CategoryID
		}
		if input.Slug != nil {
			guide.Slug = *input.Slug
		}
		if input.Icon != nil {
			guide.Icon = input.Icon
		}
		if input.SortOrder != nil {
			guide.SortOrder = *input.SortOrder
		}
		if err := u.guideRepo.Update(txCtx, guide); err != nil {
			return err
		}
		if input.Translations != nil {
			if err := u.setGuideTranslations(txCtx, id, input.Translations); err != nil {
				return err
			}
		}
		if input.Conditions != nil {
			return u.setGuideConditions(txCtx, id, input.Conditions)
		}
		return nil
	})
}

func (u *guideAdminUsecase) DeleteGuide(ctx context.Context, id uuid.UUID) error {
	return u.guideRepo.Delete(ctx, id)
}

func (u *guideAdminUsecase) AddGuideCondition(ctx context.Context, guideID uuid.UUID, cond usecase.ConditionInput) error {
	condition := &entity.GuideCondition{
		GuideID:        guideID,
		ConditionType:  cond.ConditionType,
		Operator:       cond.Operator,
		ConditionValue: toJSONMap(cond.ConditionValue),
		IsInverse:      cond.IsInverse,
	}
	return u.guideRepo.AddCondition(ctx, condition)
}

func (u *guideAdminUsecase) RemoveGuideCondition(ctx context.Context, condID uuid.UUID) error {
	return u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := u.guideRepo.RemoveCondition(txCtx, condID); err != nil {
			return err
		}
		return nil
	})
}

func (u *guideAdminUsecase) SetGuideTranslations(ctx context.Context, guideID uuid.UUID, translations []usecase.TranslationInput) error {
	return u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		return u.setGuideTranslations(txCtx, guideID, translations)
	})
}

func (u *guideAdminUsecase) CreateStep(ctx context.Context, input usecase.CreateStepInput) (*entity.GuideStep, error) {
	var step *entity.GuideStep
	err := u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		maxOrder, err := u.stepRepo.GetMaxSortOrder(txCtx, input.GuideID)
		if err != nil {
			return err
		}
		step = &entity.GuideStep{
			GuideID:         input.GuideID,
			Slug:            input.Slug,
			StepType:        input.StepType,
			SortOrder:       maxOrder + 1,
			IsOptional:      input.IsOptional,
			EstimatedTime:   input.EstimatedTime,
			DifficultyLevel: input.DifficultyLevel,
			FeeEstimate:     input.FeeEstimate,
		}
		if input.EffectiveDate != nil {
			step.EffectiveDate = *input.EffectiveDate
		}
		if input.ExpiryDate != nil {
			step.ExpiryDate = input.ExpiryDate
		}
		if err := u.stepRepo.Create(txCtx, step); err != nil {
			return err
		}
		if err := u.setStepTranslations(txCtx, step.ID, input.Translations); err != nil {
			return err
		}
		if err := u.setStepConditions(txCtx, step.ID, input.Conditions); err != nil {
			return err
		}
		return u.setStepDependencies(txCtx, step.ID, input.Dependencies)
	})
	if err != nil {
		return nil, err
	}
	return step, nil
}

func (u *guideAdminUsecase) UpdateStep(ctx context.Context, id uuid.UUID, input usecase.UpdateStepInput) error {
	return u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		step, err := u.stepRepo.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		if input.Slug != nil {
			step.Slug = *input.Slug
		}
		if input.StepType != nil {
			step.StepType = *input.StepType
		}
		if input.SortOrder != nil {
			step.SortOrder = *input.SortOrder
		}
		if input.IsOptional != nil {
			step.IsOptional = *input.IsOptional
		}
		if input.EstimatedTime != nil {
			step.EstimatedTime = input.EstimatedTime
		}
		if input.DifficultyLevel != nil {
			step.DifficultyLevel = input.DifficultyLevel
		}
		if input.FeeEstimate != nil {
			step.FeeEstimate = input.FeeEstimate
		}
		if input.EffectiveDate != nil {
			step.EffectiveDate = *input.EffectiveDate
		}
		if input.ExpiryDate != nil {
			step.ExpiryDate = input.ExpiryDate
		}
		if err := u.stepRepo.Update(txCtx, step); err != nil {
			return err
		}
		if input.Translations != nil {
			if err := u.setStepTranslations(txCtx, id, input.Translations); err != nil {
				return err
			}
		}
		if input.Conditions != nil {
			if err := u.setStepConditions(txCtx, id, input.Conditions); err != nil {
				return err
			}
		}
		if input.Dependencies != nil {
			if err := u.setStepDependencies(txCtx, id, input.Dependencies); err != nil {
				return err
			}
		}
		if input.Translations != nil || input.Conditions != nil || input.Dependencies != nil || input.Slug != nil || input.StepType != nil || input.SortOrder != nil || input.IsOptional != nil || input.EstimatedTime != nil || input.DifficultyLevel != nil || input.FeeEstimate != nil || input.EffectiveDate != nil || input.ExpiryDate != nil {
			return u.progressRepo.InvalidateJourneysForGuide(txCtx, step.GuideID)
		}
		return nil
	})
}

func (u *guideAdminUsecase) DeleteStep(ctx context.Context, id uuid.UUID) error {
	return u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		step, err := u.stepRepo.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		if err := u.stepRepo.HardDelete(txCtx, id); err != nil {
			return err
		}
		return u.progressRepo.InvalidateJourneysForGuide(txCtx, step.GuideID)
	})
}

func (u *guideAdminUsecase) ReorderSteps(ctx context.Context, guideID uuid.UUID, stepIDsInOrder []uuid.UUID) error {
	return u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := u.stepRepo.Reorder(txCtx, guideID, stepIDsInOrder); err != nil {
			return err
		}
		return u.progressRepo.InvalidateJourneysForGuide(txCtx, guideID)
	})
}

func (u *guideAdminUsecase) AddStepCondition(ctx context.Context, stepID uuid.UUID, cond usecase.ConditionInput) error {
	return u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		condition := &entity.StepCondition{
			StepID:         stepID,
			ConditionType:  cond.ConditionType,
			Operator:       cond.Operator,
			ConditionValue: toJSONMap(cond.ConditionValue),
			IsInverse:      cond.IsInverse,
		}
		if err := u.stepRepo.AddCondition(txCtx, condition); err != nil {
			return err
		}
		return u.invalidateJourneysForStep(txCtx, stepID)
	})
}

func (u *guideAdminUsecase) RemoveStepCondition(ctx context.Context, condID uuid.UUID) error {
	return u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		cond, err := u.stepRepo.GetCondition(txCtx, condID)
		if err != nil {
			return err
		}
		if err := u.stepRepo.RemoveCondition(txCtx, condID); err != nil {
			return err
		}
		return u.invalidateJourneysForStep(txCtx, cond.StepID)
	})
}

func (u *guideAdminUsecase) AddStepDependency(ctx context.Context, stepID, requiredStepID uuid.UUID, depType entity.DependencyType) error {
	return u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if stepID == requiredStepID {
			return apperrors.BadRequestError("guide.errors.dependencyCycle")
		}
		if err := u.ensureNoDependencyCycle(txCtx, stepID, requiredStepID); err != nil {
			return err
		}
		dependency := &entity.StepDependency{
			StepID:         stepID,
			RequiredStepID: requiredStepID,
			DependencyType: depType,
		}
		if err := u.stepRepo.AddDependency(txCtx, dependency); err != nil {
			return err
		}
		return u.invalidateJourneysForStep(txCtx, stepID)
	})
}

func (u *guideAdminUsecase) RemoveStepDependency(ctx context.Context, depID uuid.UUID) error {
	return u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		dep, err := u.stepRepo.GetDependency(txCtx, depID)
		if err != nil {
			return err
		}
		if err := u.stepRepo.RemoveDependency(txCtx, depID); err != nil {
			return err
		}
		return u.invalidateJourneysForStep(txCtx, dep.StepID)
	})
}

func (u *guideAdminUsecase) SetStepTranslations(ctx context.Context, stepID uuid.UUID, translations []usecase.StepTranslationInput) error {
	return u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		return u.setStepTranslations(txCtx, stepID, translations)
	})
}

func (u *guideAdminUsecase) GetStepVersions(ctx context.Context, stepID uuid.UUID, q query.QueryOptions) ([]*entity.GuideStepVersion, error) {
	return u.stepRepo.GetVersions(ctx, stepID, q)
}

func (u *guideAdminUsecase) RevertStepToVersion(ctx context.Context, stepID uuid.UUID, version int) error {
	_ = stepID
	_ = version
	return guideerror.ErrVersionRestoreNotSupported
}

func (u *guideAdminUsecase) setCategoryTranslations(ctx context.Context, categoryID uuid.UUID, translations []usecase.TranslationInput) error {
	current, err := u.catRepo.GetTranslations(ctx, categoryID)
	if err != nil {
		return err
	}
	incoming := make(map[string]usecase.TranslationInput, len(translations))
	for _, t := range translations {
		if t.Language == "" {
			continue
		}
		incoming[t.Language] = t
		if err := u.catRepo.UpsertTranslation(ctx, &entity.GuideCategoryTranslation{
			GuideCategoryID: categoryID,
			Language:        t.Language,
			Name:            t.Name,
			Description:     t.Description,
		}); err != nil {
			return err
		}
	}
	for _, existing := range current {
		if _, ok := incoming[existing.Language]; !ok {
			if err := u.catRepo.DeleteTranslation(ctx, categoryID, existing.Language); err != nil {
				return err
			}
		}
	}
	return nil
}

func (u *guideAdminUsecase) setGuideTranslations(ctx context.Context, guideID uuid.UUID, translations []usecase.TranslationInput) error {
	current, err := u.guideRepo.GetTranslations(ctx, guideID)
	if err != nil {
		return err
	}
	incoming := make(map[string]usecase.TranslationInput, len(translations))
	for _, t := range translations {
		if t.Language == "" {
			continue
		}
		incoming[t.Language] = t
		if err := u.guideRepo.UpsertTranslation(ctx, &entity.GuideTranslation{
			GuideID:     guideID,
			Language:    t.Language,
			Name:        t.Name,
			Description: t.Description,
		}); err != nil {
			return err
		}
	}
	for _, existing := range current {
		if _, ok := incoming[existing.Language]; !ok {
			if err := u.guideRepo.DeleteTranslation(ctx, guideID, existing.Language); err != nil {
				return err
			}
		}
	}
	return nil
}

func (u *guideAdminUsecase) setStepTranslations(ctx context.Context, stepID uuid.UUID, translations []usecase.StepTranslationInput) error {
	current, err := u.stepRepo.GetTranslations(ctx, stepID)
	if err != nil {
		return err
	}
	incoming := make(map[string]usecase.StepTranslationInput, len(translations))
	for _, t := range translations {
		if t.Language == "" {
			continue
		}
		incoming[t.Language] = t
		if err := u.stepRepo.UpsertTranslation(ctx, &entity.GuideStepTranslation{
			GuideStepID:     stepID,
			Language:        t.Language,
			Title:           t.Title,
			Description:     t.Description,
			DetailedContent: toJSONMap(t.DetailedContent),
		}); err != nil {
			return err
		}
	}
	for _, existing := range current {
		if _, ok := incoming[existing.Language]; !ok {
			if err := u.stepRepo.DeleteTranslation(ctx, stepID, existing.Language); err != nil {
				return err
			}
		}
	}
	return nil
}

func (u *guideAdminUsecase) setCategoryConditions(ctx context.Context, categoryID uuid.UUID, conditions []usecase.ConditionInput) error {
	if err := u.removeCategoryConditions(ctx, categoryID); err != nil {
		return err
	}
	for _, cond := range conditions {
		if err := u.catRepo.AddCondition(ctx, &entity.GuideCategoryCondition{
			CategoryID:     categoryID,
			ConditionType:  cond.ConditionType,
			Operator:       cond.Operator,
			ConditionValue: toJSONMap(cond.ConditionValue),
			IsInverse:      cond.IsInverse,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (u *guideAdminUsecase) setGuideConditions(ctx context.Context, guideID uuid.UUID, conditions []usecase.ConditionInput) error {
	if err := u.removeGuideConditions(ctx, guideID); err != nil {
		return err
	}
	for _, cond := range conditions {
		if err := u.guideRepo.AddCondition(ctx, &entity.GuideCondition{
			GuideID:        guideID,
			ConditionType:  cond.ConditionType,
			Operator:       cond.Operator,
			ConditionValue: toJSONMap(cond.ConditionValue),
			IsInverse:      cond.IsInverse,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (u *guideAdminUsecase) setStepConditions(ctx context.Context, stepID uuid.UUID, conditions []usecase.ConditionInput) error {
	if err := u.removeStepConditions(ctx, stepID); err != nil {
		return err
	}
	for _, cond := range conditions {
		if err := u.stepRepo.AddCondition(ctx, &entity.StepCondition{
			StepID:         stepID,
			ConditionType:  cond.ConditionType,
			Operator:       cond.Operator,
			ConditionValue: toJSONMap(cond.ConditionValue),
			IsInverse:      cond.IsInverse,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (u *guideAdminUsecase) setStepDependencies(ctx context.Context, stepID uuid.UUID, deps []usecase.DependencyInput) error {
	if err := u.removeStepDependencies(ctx, stepID); err != nil {
		return err
	}
	for _, dep := range deps {
		if stepID == dep.RequiredStepID {
			return apperrors.BadRequestError("guide.errors.dependencyCycle")
		}
		if err := u.ensureNoDependencyCycle(ctx, stepID, dep.RequiredStepID); err != nil {
			return err
		}
		if err := u.stepRepo.AddDependency(ctx, &entity.StepDependency{
			StepID:         stepID,
			RequiredStepID: dep.RequiredStepID,
			DependencyType: dep.DependencyType,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (u *guideAdminUsecase) removeCategoryConditions(ctx context.Context, categoryID uuid.UUID) error {
	conditions, err := u.catRepo.GetConditions(ctx, categoryID)
	if err != nil {
		return err
	}
	for _, cond := range conditions {
		if err := u.catRepo.RemoveCondition(ctx, cond.ID); err != nil && err != guideerror.ErrCategoryConditionNotFound {
			return err
		}
	}
	return nil
}

func (u *guideAdminUsecase) removeGuideConditions(ctx context.Context, guideID uuid.UUID) error {
	conditions, err := u.guideRepo.GetConditions(ctx, guideID)
	if err != nil {
		return err
	}
	for _, cond := range conditions {
		if err := u.guideRepo.RemoveCondition(ctx, cond.ID); err != nil && err != guideerror.ErrGuideConditionNotFound {
			return err
		}
	}
	return nil
}

func (u *guideAdminUsecase) removeStepConditions(ctx context.Context, stepID uuid.UUID) error {
	conditions, err := u.stepRepo.GetConditions(ctx, stepID)
	if err != nil {
		return err
	}
	for _, cond := range conditions {
		if err := u.stepRepo.RemoveCondition(ctx, cond.ID); err != nil && err != guideerror.ErrStepConditionNotFound {
			return err
		}
	}
	return nil
}

func (u *guideAdminUsecase) removeStepDependencies(ctx context.Context, stepID uuid.UUID) error {
	deps, err := u.stepRepo.GetDependencies(ctx, stepID)
	if err != nil {
		return err
	}
	for _, dep := range deps {
		if err := u.stepRepo.RemoveDependency(ctx, dep.ID); err != nil && err != guideerror.ErrDependencyNotFound {
			return err
		}
	}
	return nil
}

func toJSONMap(value interface{}) datatypes.JSONMap {
	if value == nil {
		return datatypes.JSONMap{}
	}
	if jsonMap, ok := value.(datatypes.JSONMap); ok {
		return jsonMap
	}
	switch v := value.(type) {
	case map[string]interface{}:
		return datatypes.JSONMap(v)
	case map[string]string:
		result := datatypes.JSONMap{}
		for key, val := range v {
			result[key] = val
		}
		return result
	default:
		return datatypes.JSONMap{"value": v}
	}
}

func (u *guideAdminUsecase) ensureNoDependencyCycle(ctx context.Context, stepID, requiredStepID uuid.UUID) error {
	visited := map[uuid.UUID]bool{}
	queue := []uuid.UUID{requiredStepID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current == stepID {
			return apperrors.BadRequestError("guide.errors.dependencyCycle")
		}
		if visited[current] {
			continue
		}
		visited[current] = true
		deps, err := u.stepRepo.GetDependencies(ctx, current)
		if err != nil {
			if errors.Is(err, guideerror.ErrStepNotFound) {
				continue
			}
			return err
		}
		for _, dep := range deps {
			queue = append(queue, dep.RequiredStepID)
		}
	}
	return nil
}

func (u *guideAdminUsecase) invalidateJourneysForStep(ctx context.Context, stepID uuid.UUID) error {
	step, err := u.stepRepo.GetByID(ctx, stepID)
	if err != nil {
		return err
	}
	return u.progressRepo.InvalidateJourneysForGuide(ctx, step.GuideID)
}
