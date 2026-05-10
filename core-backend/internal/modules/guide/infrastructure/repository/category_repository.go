package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/entity"
	guideerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/error"
	guiderepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type categoryRepository struct {
	sharedrepo.GenericRepository[entity.GuideCategory]
	db     *core.Database
	logger core.Logger
}

func NewCategoryRepository(db *core.Database, logger core.Logger) guiderepo.CategoryRepository {
	base := sharedrepo.NewBaseRepository[entity.GuideCategory](db, logger)
	return &categoryRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *categoryRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *categoryRepository) GetBySlug(ctx context.Context, parentID *uuid.UUID, slug string, locale constants.Locale) (*entity.GuideCategory, error) {
	var category entity.GuideCategory
	db := r.getDB(ctx)

	baseQuery := db.Where("slug = ?", slug)
	if parentID == nil {
		baseQuery = baseQuery.Where("parent_category_id IS NULL")
	} else {
		baseQuery = baseQuery.Where("parent_category_id = ?", *parentID)
	}

	if err := baseQuery.Preload("Translations", "language = ?", locale).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, guideerror.ErrCategoryNotFound
		}
		r.logger.Error("Failed to get category by slug", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	if len(category.Translations) == 0 {
		if err := db.Preload("Translations", "language = ?", constants.LocaleEnglish).
			Where("id = ?", category.ID).
			First(&category).Error; err != nil {
			return nil, errors.InternalError("errors.databaseError", err)
		}
	}

	return &category, nil
}

func (r *categoryRepository) ListTree(ctx context.Context, includeInactive bool, locale constants.Locale) ([]*entity.GuideCategory, error) {
	var categories []*entity.GuideCategory
	db := r.getDB(ctx).
		Preload("Translations", "language = ?", locale).
		Order("sort_order asc, created_at asc")
	// TODO: apply active/inactive filtering once publish-state exists in the guide domain.
	_ = includeInactive
	if err := db.Find(&categories).Error; err != nil {
		r.logger.Error("Failed to list category tree", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	for i := range categories {
		if len(categories[i].Translations) == 0 {
			if err := r.getDB(ctx).
				Preload("Translations", "language = ?", constants.LocaleEnglish).
				First(categories[i], "id = ?", categories[i].ID).Error; err != nil {
				r.logger.Error("Failed to load fallback translation for category", core.Error(err))
			}
		}
	}

	return categories, nil
}

func (r *categoryRepository) GetConditions(ctx context.Context, categoryID uuid.UUID) ([]*entity.GuideCategoryCondition, error) {
	var conditions []*entity.GuideCategoryCondition
	if err := r.getDB(ctx).Where("category_id = ?", categoryID).Order("created_at asc").Find(&conditions).Error; err != nil {
		r.logger.Error("Failed to get category conditions", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return conditions, nil
}

func (r *categoryRepository) AddCondition(ctx context.Context, cond *entity.GuideCategoryCondition) error {
	if err := r.getDB(ctx).Create(cond).Error; err != nil {
		r.logger.Error("Failed to add category condition", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *categoryRepository) RemoveCondition(ctx context.Context, condID uuid.UUID) error {
	result := r.getDB(ctx).Delete(&entity.GuideCategoryCondition{}, "id = ?", condID)
	if result.Error != nil {
		r.logger.Error("Failed to remove category condition", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return guideerror.ErrCategoryConditionNotFound
	}
	return nil
}

func (r *categoryRepository) GetTranslations(ctx context.Context, categoryID uuid.UUID) ([]*entity.GuideCategoryTranslation, error) {
	var translations []*entity.GuideCategoryTranslation
	if err := r.getDB(ctx).Where("guide_category_id = ?", categoryID).Order("language asc").Find(&translations).Error; err != nil {
		r.logger.Error("Failed to get category translations", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return translations, nil
}

func (r *categoryRepository) UpsertTranslation(ctx context.Context, trans *entity.GuideCategoryTranslation) error {
	if err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "guide_category_id"}, {Name: "language"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"name":        trans.Name,
			"description": trans.Description,
			"updated_at":  gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(trans).Error; err != nil {
		r.logger.Error("Failed to upsert category translation", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *categoryRepository) DeleteTranslation(ctx context.Context, categoryID uuid.UUID, language string) error {
	result := r.getDB(ctx).Where("guide_category_id = ? AND language = ?", categoryID, language).Delete(&entity.GuideCategoryTranslation{})
	if result.Error != nil {
		r.logger.Error("Failed to delete category translation", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return guideerror.ErrTranslationNotFound
	}
	return nil
}
