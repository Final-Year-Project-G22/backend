package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/entity"
	guideerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/error"
	guiderepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type stepRepository struct {
	sharedrepo.GenericRepository[entity.GuideStep]
	db     *core.Database
	logger core.Logger
}

func NewStepRepository(db *core.Database, logger core.Logger) guiderepo.StepRepository {
	base := sharedrepo.NewBaseRepository[entity.GuideStep](db, logger)
	return &stepRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *stepRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *stepRepository) GetBySlug(ctx context.Context, guideID uuid.UUID, slug string, locale constants.Locale) (*entity.GuideStep, error) {
	var step entity.GuideStep
	if err := r.getDB(ctx).Preload("Translations", "language = ?", locale).Where("guide_id = ? AND slug = ?", guideID, slug).First(&step).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, guideerror.ErrStepNotFound
		}
		r.logger.Error("Failed to get step by slug", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	if len(step.Translations) == 0 {
		if err := r.getDB(ctx).Preload("Translations", "language = ?", constants.LocaleEnglish).
			Where("id = ?", step.ID).
			First(&step).Error; err != nil {
			return nil, errors.InternalError("errors.databaseError", err)
		}
	}
	return &step, nil
}

func (r *stepRepository) ListByGuide(ctx context.Context, guideID uuid.UUID, q query.QueryOptions, locale constants.Locale) ([]*entity.GuideStep, error) {
	var steps []*entity.GuideStep
	db := r.getDB(ctx).Where("guide_id = ?", guideID)
	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	if len(q.Preload) == 0 {
		db = db.Preload("Translations", "language = ?", locale)
	}
	if q.Search != "" {
		db = db.Where("slug ILIKE ?", "%"+q.Search+"%")
	}
	if len(q.SortBy) > 0 {
		for i, col := range q.SortBy {
			order := "asc"
			if i < len(q.SortOrder) && q.SortOrder[i] == "desc" {
				order = "desc"
			}
			db = db.Order(fmt.Sprintf("%s %s", col, order))
		}
	} else {
		db = db.Order("sort_order asc, created_at asc")
	}
	if q.Page < 1 {
		q.Page = query.DefaultPage
	}
	if q.PageSize < 1 {
		q.PageSize = query.DefaultPageSize
	}
	if q.PageSize > query.MaxPageSize {
		q.PageSize = query.MaxPageSize
	}
	if err := db.Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&steps).Error; err != nil {
		r.logger.Error("Failed to list steps by guide", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	for i := range steps {
		if len(steps[i].Translations) == 0 {
			if err := r.getDB(ctx).Preload("Translations", "language = ?", constants.LocaleEnglish).
				Where("id = ?", steps[i].ID).
				First(steps[i]).Error; err != nil {
				r.logger.Error("Failed to load fallback translation for step", core.Error(err))
			}
		}
	}

	return steps, nil
}

func (r *stepRepository) Reorder(ctx context.Context, guideID uuid.UUID, stepIDsInOrder []uuid.UUID) error {
	if len(stepIDsInOrder) == 0 {
		return nil
	}
	db := r.getDB(ctx)

	// Phase 0: Lock all active steps for this guide to prevent concurrent modifications
	var lockedIDs []uuid.UUID
	if err := db.Raw("SELECT id FROM guide_steps WHERE guide_id = ? AND deleted_at IS NULL FOR UPDATE", guideID).Scan(&lockedIDs).Error; err != nil {
		r.logger.Error("Failed to lock steps for reorder", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}

	// Validate that the frontend sent ALL active steps (not a partial list)
	if len(lockedIDs) != len(stepIDsInOrder) {
		return guideerror.ErrIncompleteStepList
	}

	// Hard-delete any legacy soft-deleted rows for this guide.
	if err := db.Exec("DELETE FROM guide_steps WHERE guide_id = ? AND deleted_at IS NOT NULL", guideID).Error; err != nil {
		r.logger.Error("Failed to cleanup soft-deleted steps", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}

	// Phase 1: Move all active rows to unique temporary negative values using a CTE.
	// This frees up their current sort_order values so Phase 2 can assign
	// the final 1-based ordering without violating the unique constraint.
	shiftSQL := `WITH temp AS (
		SELECT id, -(ROW_NUMBER() OVER (ORDER BY sort_order, created_at)) AS new_order
		FROM guide_steps
		WHERE guide_id = ? AND deleted_at IS NULL
	)
	UPDATE guide_steps gs SET sort_order = temp.new_order, updated_at = NOW()
	FROM temp WHERE gs.id = temp.id`
	if err := db.Exec(shiftSQL, guideID).Error; err != nil {
		r.logger.Error("Failed to shift step sort orders to temp values", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}

	// Phase 2: assign final consecutive sort_order (1-based) using parameterized CASE.
	// Integer THEN values are embedded as literals (safe: sequential counters, not user input).
	// UUID WHEN values and IN clause use ? placeholders.
	caseSQL := "UPDATE guide_steps SET sort_order = CAST(CASE id "
	args := make([]interface{}, 0, len(stepIDsInOrder)+1)
	for i, stepID := range stepIDsInOrder {
		caseSQL += fmt.Sprintf("WHEN ? THEN %d ", i+1)
		args = append(args, stepID)
	}
	caseSQL += "END AS integer), updated_at = NOW() WHERE id IN ("
	inPlaceholders := make([]string, len(stepIDsInOrder))
	for i, stepID := range stepIDsInOrder {
		inPlaceholders[i] = "?"
		args = append(args, stepID)
	}
	caseSQL += strings.Join(inPlaceholders, ",") + ") AND guide_id = ?"
	args = append(args, guideID)

	if err := db.Exec(caseSQL, args...).Error; err != nil {
		r.logger.Error("Failed to set final step sort order", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}

	return nil
}

func (r *stepRepository) GetMaxSortOrder(ctx context.Context, guideID uuid.UUID) (int, error) {
	var maxOrder int
	if err := r.getDB(ctx).Model(&entity.GuideStep{}).
		Select("COALESCE(MAX(sort_order), 0)").
		Where("guide_id = ? AND deleted_at IS NULL", guideID).
		Scan(&maxOrder).Error; err != nil {
		r.logger.Error("Failed to get max sort order", core.Error(err))
		return 0, errors.InternalError("errors.databaseError", err)
	}
	return maxOrder, nil
}

func (r *stepRepository) GetConditions(ctx context.Context, stepID uuid.UUID) ([]*entity.StepCondition, error) {
	var conditions []*entity.StepCondition
	if err := r.getDB(ctx).Where("step_id = ?", stepID).Order("created_at asc").Find(&conditions).Error; err != nil {
		r.logger.Error("Failed to get step conditions", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return conditions, nil
}

func (r *stepRepository) GetCondition(ctx context.Context, condID uuid.UUID) (*entity.StepCondition, error) {
	var condition entity.StepCondition
	if err := r.getDB(ctx).Where("id = ?", condID).First(&condition).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, guideerror.ErrStepConditionNotFound
		}
		r.logger.Error("Failed to get step condition", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &condition, nil
}

func (r *stepRepository) AddCondition(ctx context.Context, cond *entity.StepCondition) error {
	if err := r.getDB(ctx).Create(cond).Error; err != nil {
		r.logger.Error("Failed to add step condition", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *stepRepository) RemoveCondition(ctx context.Context, condID uuid.UUID) error {
	result := r.getDB(ctx).Delete(&entity.StepCondition{}, "id = ?", condID)
	if result.Error != nil {
		r.logger.Error("Failed to remove step condition", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return guideerror.ErrStepConditionNotFound
	}
	return nil
}

func (r *stepRepository) GetDependencies(ctx context.Context, stepID uuid.UUID) ([]*entity.StepDependency, error) {
	var dependencies []*entity.StepDependency
	if err := r.getDB(ctx).Preload("RequiredStep").Where("step_id = ?", stepID).Order("created_at asc").Find(&dependencies).Error; err != nil {
		r.logger.Error("Failed to get step dependencies", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return dependencies, nil
}

func (r *stepRepository) GetDependency(ctx context.Context, depID uuid.UUID) (*entity.StepDependency, error) {
	var dependency entity.StepDependency
	if err := r.getDB(ctx).Preload("RequiredStep").Where("id = ?", depID).First(&dependency).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, guideerror.ErrDependencyNotFound
		}
		r.logger.Error("Failed to get step dependency", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &dependency, nil
}

func (r *stepRepository) AddDependency(ctx context.Context, dep *entity.StepDependency) error {
	if err := r.getDB(ctx).Create(dep).Error; err != nil {
		r.logger.Error("Failed to add step dependency", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *stepRepository) RemoveDependency(ctx context.Context, depID uuid.UUID) error {
	result := r.getDB(ctx).Delete(&entity.StepDependency{}, "id = ?", depID)
	if result.Error != nil {
		r.logger.Error("Failed to remove step dependency", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return guideerror.ErrDependencyNotFound
	}
	return nil
}

func (r *stepRepository) GetTranslations(ctx context.Context, stepID uuid.UUID) ([]*entity.GuideStepTranslation, error) {
	var translations []*entity.GuideStepTranslation
	if err := r.getDB(ctx).Where("guide_step_id = ?", stepID).Order("language asc").Find(&translations).Error; err != nil {
		r.logger.Error("Failed to get step translations", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return translations, nil
}

func (r *stepRepository) UpsertTranslation(ctx context.Context, trans *entity.GuideStepTranslation) error {
	if err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "guide_step_id"}, {Name: "language"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":            trans.Title,
			"description":      trans.Description,
			"detailed_content": trans.DetailedContent,
			"updated_at":       gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(trans).Error; err != nil {
		r.logger.Error("Failed to upsert step translation", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *stepRepository) DeleteTranslation(ctx context.Context, stepID uuid.UUID, language string) error {
	result := r.getDB(ctx).Where("guide_step_id = ? AND language = ?", stepID, language).Delete(&entity.GuideStepTranslation{})
	if result.Error != nil {
		r.logger.Error("Failed to delete step translation", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return guideerror.ErrTranslationNotFound
	}
	return nil
}

func (r *stepRepository) GetVersions(ctx context.Context, stepID uuid.UUID, q query.QueryOptions) ([]*entity.GuideStepVersion, error) {
	var versions []*entity.GuideStepVersion
	db := r.getDB(ctx).Where("step_id = ?", stepID)
	if len(q.SortBy) > 0 {
		for i, col := range q.SortBy {
			order := "asc"
			if i < len(q.SortOrder) && q.SortOrder[i] == "desc" {
				order = "desc"
			}
			db = db.Order(fmt.Sprintf("%s %s", col, order))
		}
	} else {
		db = db.Order("version desc")
	}
	if q.Page < 1 {
		q.Page = query.DefaultPage
	}
	if q.PageSize < 1 {
		q.PageSize = query.DefaultPageSize
	}
	if q.PageSize > query.MaxPageSize {
		q.PageSize = query.MaxPageSize
	}
	if err := db.Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&versions).Error; err != nil {
		r.logger.Error("Failed to get step versions", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return versions, nil
}

func (r *stepRepository) GetVersion(ctx context.Context, stepID uuid.UUID, version int) (*entity.GuideStepVersion, error) {
	var stepVersion entity.GuideStepVersion
	if err := r.getDB(ctx).Where("step_id = ? AND version = ?", stepID, version).First(&stepVersion).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, guideerror.ErrStepVersionNotFound
		}
		r.logger.Error("Failed to get step version", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &stepVersion, nil
}

func (r *stepRepository) RestoreVersion(ctx context.Context, stepID uuid.UUID, version int) error {
	_, err := r.GetVersion(ctx, stepID, version)
	if err != nil {
		return err
	}
	return guideerror.ErrVersionRestoreNotSupported
}
