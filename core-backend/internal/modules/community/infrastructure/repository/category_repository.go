package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	communityerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/error"
	communityrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type communityCategoryRepository struct {
	sharedrepo.GenericRepository[entity.CommunityCategory]
	db     *core.Database
	logger core.Logger
}

func NewCommunityCategoryRepository(db *core.Database, logger core.Logger) communityrepo.CommunityCategoryRepository {
	base := sharedrepo.NewBaseRepository[entity.CommunityCategory](db, logger)
	return &communityCategoryRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *communityCategoryRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *communityCategoryRepository) GetBySlug(ctx context.Context, parentID *uuid.UUID, slug string) (*entity.CommunityCategory, error) {
	var category entity.CommunityCategory
	baseQuery := r.getDB(ctx).Where("slug = ?", slug)
	if parentID == nil {
		baseQuery = baseQuery.Where("parent_category_id IS NULL")
	} else {
		baseQuery = baseQuery.Where("parent_category_id = ?", *parentID)
	}
	if err := baseQuery.First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, communityerror.ErrCategoryNotFound
		}
		r.logger.Error("Failed to get community category by slug", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &category, nil
}

func (r *communityCategoryRepository) ListTree(ctx context.Context, includeInactive bool) ([]*entity.CommunityCategory, error) {
	var categories []*entity.CommunityCategory
	db := r.getDB(ctx).Order("created_at asc")
	if !includeInactive {
		db = db.Where("is_active = ?", true)
	}
	if err := db.Find(&categories).Error; err != nil {
		r.logger.Error("Failed to list community category tree", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return categories, nil
}

func (r *communityCategoryRepository) ListActive(ctx context.Context, q query.QueryOptions) ([]*entity.CommunityCategory, error) {
	var categories []*entity.CommunityCategory
	db := r.getDB(ctx).Where("is_active = ?", true)
	if q.Search != "" {
		search := "%" + q.Search + "%"
		db = db.Where("name ILIKE ? OR slug ILIKE ?", search, search)
	}
	db = applyPaginationAndSorting(db, q, "created_at asc")
	if err := db.Find(&categories).Error; err != nil {
		r.logger.Error("Failed to list active community categories", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return categories, nil
}

func (r *communityCategoryRepository) ExistsBySlug(ctx context.Context, parentID *uuid.UUID, slug string) (bool, error) {
	var count int64
	db := r.getDB(ctx).Model(&entity.CommunityCategory{}).Where("slug = ?", slug)
	if parentID == nil {
		db = db.Where("parent_category_id IS NULL")
	} else {
		db = db.Where("parent_category_id = ?", *parentID)
	}
	if err := db.Count(&count).Error; err != nil {
		r.logger.Error("Failed to check category slug existence", core.Error(err))
		return false, errors.InternalError("errors.databaseError", err)
	}
	return count > 0, nil
}
