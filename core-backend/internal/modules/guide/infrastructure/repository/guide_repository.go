package repository

import (
	"context"
	"fmt"

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

type guideRepository struct {
	sharedrepo.GenericRepository[entity.Guide]
	db     *core.Database
	logger core.Logger
}

func NewGuideRepository(db *core.Database, logger core.Logger) guiderepo.GuideRepository {
	base := sharedrepo.NewBaseRepository[entity.Guide](db, logger)
	return &guideRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *guideRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *guideRepository) GetBySlug(ctx context.Context, categoryID uuid.UUID, slug string, locale constants.Locale) (*entity.Guide, error) {
	var guide entity.Guide
	if err := r.getDB(ctx).Preload("Translations", "language = ?", locale).Where("category_id = ? AND slug = ?", categoryID, slug).First(&guide).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, guideerror.ErrGuideNotFound
		}
		r.logger.Error("Failed to get guide by slug", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	if len(guide.Translations) == 0 {
		if err := r.getDB(ctx).Preload("Translations", "language = ?", constants.LocaleEnglish).
			Where("id = ?", guide.ID).
			First(&guide).Error; err != nil {
			return nil, errors.InternalError("errors.databaseError", err)
		}
	}
	return &guide, nil
}

func (r *guideRepository) GetBySlugGlobal(ctx context.Context, slug string, locale constants.Locale) (*entity.Guide, error) {
	var guide entity.Guide
	if err := r.getDB(ctx).Preload("Translations", "language = ?", locale).Where("slug = ?", slug).First(&guide).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, guideerror.ErrGuideNotFound
		}
		r.logger.Error("Failed to get guide by global slug", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	if len(guide.Translations) == 0 {
		if err := r.getDB(ctx).Preload("Translations", "language = ?", constants.LocaleEnglish).
			Where("id = ?", guide.ID).
			First(&guide).Error; err != nil {
			return nil, errors.InternalError("errors.databaseError", err)
		}
	}
	return &guide, nil
}

func (r *guideRepository) ListByCategory(ctx context.Context, categoryID uuid.UUID, q query.QueryOptions, locale constants.Locale) ([]*entity.Guide, error) {
	var guides []*entity.Guide
	db := r.getDB(ctx).Where("category_id = ?", categoryID)
	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	if len(q.Preload) == 0 {
		db = db.Preload("Translations", "language = ?", locale)
	}
	// TODO: integrate condition evaluation against business profile once guide visibility rules are implemented.
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
	if err := db.Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&guides).Error; err != nil {
		r.logger.Error("Failed to list guides by category", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	for i := range guides {
		if len(guides[i].Translations) == 0 {
			if err := r.getDB(ctx).Preload("Translations", "language = ?", constants.LocaleEnglish).
				Where("id = ?", guides[i].ID).
				First(guides[i]).Error; err != nil {
				r.logger.Error("Failed to load fallback translation for guide", core.Error(err))
			}
		}
	}

	return guides, nil
}

func (r *guideRepository) Search(ctx context.Context, keyword string, q query.QueryOptions, locale constants.Locale) ([]*entity.Guide, error) {
	var guides []*entity.Guide
	search := "%" + keyword + "%"
	db := r.getDB(ctx).
		Model(&entity.Guide{}).
		Distinct("guides.*").
		Joins("LEFT JOIN guide_translations gt ON gt.guide_id = guides.id AND gt.language = ?", locale).
		Where("guides.slug ILIKE ? OR gt.name ILIKE ? OR gt.description ILIKE ?", search, search, search).
		Preload("Translations", "language = ?", locale)
	if len(q.SortBy) > 0 {
		for i, col := range q.SortBy {
			order := "asc"
			if i < len(q.SortOrder) && q.SortOrder[i] == "desc" {
				order = "desc"
			}
			db = db.Order(fmt.Sprintf("%s %s", col, order))
		}
	} else {
		db = db.Order("guides.sort_order asc, guides.created_at asc")
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
	if err := db.Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&guides).Error; err != nil {
		r.logger.Error("Failed to search guides", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	for i := range guides {
		if len(guides[i].Translations) == 0 {
			if err := r.getDB(ctx).Preload("Translations", "language = ?", constants.LocaleEnglish).
				Where("id = ?", guides[i].ID).
				First(guides[i]).Error; err != nil {
				r.logger.Error("Failed to load fallback translation for guide", core.Error(err))
			}
		}
	}

	return guides, nil
}

func (r *guideRepository) GetConditions(ctx context.Context, guideID uuid.UUID) ([]*entity.GuideCondition, error) {
	var conditions []*entity.GuideCondition
	if err := r.getDB(ctx).Where("guide_id = ?", guideID).Order("created_at asc").Find(&conditions).Error; err != nil {
		r.logger.Error("Failed to get guide conditions", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return conditions, nil
}

func (r *guideRepository) AddCondition(ctx context.Context, cond *entity.GuideCondition) error {
	if err := r.getDB(ctx).Create(cond).Error; err != nil {
		r.logger.Error("Failed to add guide condition", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *guideRepository) RemoveCondition(ctx context.Context, condID uuid.UUID) error {
	result := r.getDB(ctx).Delete(&entity.GuideCondition{}, "id = ?", condID)
	if result.Error != nil {
		r.logger.Error("Failed to remove guide condition", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return guideerror.ErrGuideConditionNotFound
	}
	return nil
}

func (r *guideRepository) GetTranslations(ctx context.Context, guideID uuid.UUID) ([]*entity.GuideTranslation, error) {
	var translations []*entity.GuideTranslation
	if err := r.getDB(ctx).Where("guide_id = ?", guideID).Order("language asc").Find(&translations).Error; err != nil {
		r.logger.Error("Failed to get guide translations", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return translations, nil
}

func (r *guideRepository) UpsertTranslation(ctx context.Context, trans *entity.GuideTranslation) error {
	if err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "guide_id"}, {Name: "language"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"name":        trans.Name,
			"description": trans.Description,
			"updated_at":  gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(trans).Error; err != nil {
		r.logger.Error("Failed to upsert guide translation", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *guideRepository) DeleteTranslation(ctx context.Context, guideID uuid.UUID, language string) error {
	result := r.getDB(ctx).Where("guide_id = ? AND language = ?", guideID, language).Delete(&entity.GuideTranslation{})
	if result.Error != nil {
		r.logger.Error("Failed to delete guide translation", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return guideerror.ErrTranslationNotFound
	}
	return nil
}
