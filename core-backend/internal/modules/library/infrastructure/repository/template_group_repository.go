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
)

type libraryTemplateGroupRepository struct {
	sharedrepo.GenericRepository[entity.LibraryTemplateGroup]
	db     *core.Database
	logger core.Logger
}

func NewLibraryTemplateGroupRepository(db *core.Database, logger core.Logger) libraryrepo.LibraryTemplateGroupRepository {
	base := sharedrepo.NewBaseRepository[entity.LibraryTemplateGroup](db, logger)
	return &libraryTemplateGroupRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *libraryTemplateGroupRepository) GetBySlug(ctx context.Context, categoryID *uuid.UUID, slug string) (*entity.LibraryTemplateGroup, error) {
	var group entity.LibraryTemplateGroup
	db := getDB(ctx, r.db).Where("slug = ?", slug)
	if categoryID != nil {
		db = db.Where("category_id = ?", *categoryID)
	}
	if err := db.Preload("Templates").First(&group).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, libraryerror.ErrTemplateGroupNotFound
		}
		r.logger.Error("Failed to get template group by slug", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &group, nil
}

func (r *libraryTemplateGroupRepository) ListByCategory(ctx context.Context, categoryID uuid.UUID, q query.QueryOptions) ([]*entity.LibraryTemplateGroup, error) {
	var groups []*entity.LibraryTemplateGroup
	db := getDB(ctx, r.db).
		Where("category_id = ? AND is_active = ?", categoryID, true).
		Preload("Templates")
	if q.Search != "" {
		db = db.Where("name ILIKE ? OR slug ILIKE ?", "%"+q.Search+"%", "%"+q.Search+"%")
	}
	db = applyPaginationAndSorting(q, "sort_order asc, created_at asc")(db)
	if err := db.Find(&groups).Error; err != nil {
		r.logger.Error("Failed to list groups by category", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return groups, nil
}

func (r *libraryTemplateGroupRepository) ListByFormat(ctx context.Context, format entity.TemplateFormat, q query.QueryOptions) ([]*entity.LibraryTemplateGroup, error) {
	var groups []*entity.LibraryTemplateGroup
	db := getDB(ctx, r.db).
		Where("format = ? AND is_active = ?", format, true).
		Preload("Templates")
	db = applyPaginationAndSorting(q, "sort_order asc, created_at asc")(db)
	if err := db.Find(&groups).Error; err != nil {
		r.logger.Error("Failed to list groups by format", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return groups, nil
}

func (r *libraryTemplateGroupRepository) IncrementDownloadCount(ctx context.Context, id uuid.UUID) error {
	result := getDB(ctx, r.db).Model(&entity.LibraryTemplateGroup{}).
		Where("id = ?", id).
		UpdateColumn("download_count", gorm.Expr("download_count + 1"))
	if result.Error != nil {
		r.logger.Error("Failed to increment download count", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return libraryerror.ErrTemplateGroupNotFound
	}
	return nil
}
