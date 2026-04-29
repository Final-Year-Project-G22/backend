package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
	libraryerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/error"
	libraryrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type libraryCategoryRepository struct {
	sharedrepo.GenericRepository[entity.LibraryCategory]
	db     *core.Database
	logger core.Logger
}

func NewLibraryCategoryRepository(db *core.Database, logger core.Logger) libraryrepo.LibraryCategoryRepository {
	base := sharedrepo.NewBaseRepository[entity.LibraryCategory](db, logger)
	return &libraryCategoryRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *libraryCategoryRepository) GetBySlug(ctx context.Context, parentID *uuid.UUID, slug string) (*entity.LibraryCategory, error) {
	var cat entity.LibraryCategory
	db := getDB(ctx, r.db).Where("slug = ?", slug)
	if parentID == nil {
		db = db.Where("parent_category_id IS NULL")
	} else {
		db = db.Where("parent_category_id = ?", *parentID)
	}
	if err := db.First(&cat).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, libraryerror.ErrCategoryNotFound
		}
		r.logger.Error("Failed to get category by slug", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &cat, nil
}

func (r *libraryCategoryRepository) ListTree(ctx context.Context, includeInactive bool) ([]*entity.LibraryCategory, error) {
	var categories []*entity.LibraryCategory
	db := getDB(ctx, r.db).Order("sort_order asc, created_at asc")
	if !includeInactive {
		db = db.Where("is_active = ?", true)
	}
	if err := db.Find(&categories).Error; err != nil {
		r.logger.Error("Failed to list category tree", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return categories, nil
}

func (r *libraryCategoryRepository) ListActive(ctx context.Context, q query.QueryOptions) ([]*entity.LibraryCategory, error) {
	var categories []*entity.LibraryCategory
	db := getDB(ctx, r.db).Where("is_active = ?", true)
	db = applyPaginationAndSorting(q, "sort_order asc, created_at asc")(db)
	if err := db.Find(&categories).Error; err != nil {
		r.logger.Error("Failed to list active categories", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return categories, nil
}

func (r *libraryCategoryRepository) GetTranslations(ctx context.Context, categoryID uuid.UUID) ([]*entity.LibraryCategoryTranslation, error) {
	var translations []*entity.LibraryCategoryTranslation
	if err := getDB(ctx, r.db).Where("library_category_id = ?", categoryID).Order("language asc").Find(&translations).Error; err != nil {
		r.logger.Error("Failed to get category translations", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return translations, nil
}

func (r *libraryCategoryRepository) UpsertTranslation(ctx context.Context, trans *entity.LibraryCategoryTranslation) error {
	if err := getDB(ctx, r.db).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "library_category_id"}, {Name: "language"}},
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

func (r *libraryCategoryRepository) DeleteTranslation(ctx context.Context, categoryID uuid.UUID, language string) error {
	result := getDB(ctx, r.db).Where("library_category_id = ? AND language = ?", categoryID, language).Delete(&entity.LibraryCategoryTranslation{})
	if result.Error != nil {
		r.logger.Error("Failed to delete category translation", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return libraryerror.ErrCategoryNotFound
	}
	return nil
}
