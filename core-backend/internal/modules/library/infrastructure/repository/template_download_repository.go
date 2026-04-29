package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
	libraryrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type libraryTemplateDownloadRepository struct {
	sharedrepo.GenericRepository[entity.LibraryTemplateDownload]
	db     *core.Database
	logger core.Logger
}

func NewLibraryTemplateDownloadRepository(db *core.Database, logger core.Logger) libraryrepo.LibraryTemplateDownloadRepository {
	base := sharedrepo.NewBaseRepository[entity.LibraryTemplateDownload](db, logger)
	return &libraryTemplateDownloadRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *libraryTemplateDownloadRepository) ListByAccount(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.LibraryTemplateDownload, error) {
	var downloads []*entity.LibraryTemplateDownload
	db := getDB(ctx, r.db).
		Where("account_id = ?", accountID).
		Order("created_at desc")
	db = applyPaginationAndSorting(q, "created_at desc")(db)
	if err := db.Find(&downloads).Error; err != nil {
		r.logger.Error("Failed to list downloads by account", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return downloads, nil
}

func (r *libraryTemplateDownloadRepository) CountByGroup(ctx context.Context, groupID uuid.UUID) (int64, error) {
	var count int64
	if err := getDB(ctx, r.db).
		Model(&entity.LibraryTemplateDownload{}).
		Where("group_id = ?", groupID).
		Count(&count).Error; err != nil {
		r.logger.Error("Failed to count downloads by group", core.Error(err))
		return 0, errors.InternalError("errors.databaseError", err)
	}
	return count, nil
}

func (r *libraryTemplateDownloadRepository) ListAll(ctx context.Context, q query.QueryOptions) ([]*entity.LibraryTemplateDownload, error) {
	var downloads []*entity.LibraryTemplateDownload
	db := getDB(ctx, r.db).Order("created_at desc")
	db = applyPaginationAndSorting(q, "created_at desc")(db)
	if err := db.Find(&downloads).Error; err != nil {
		r.logger.Error("Failed to list all downloads", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return downloads, nil
}
