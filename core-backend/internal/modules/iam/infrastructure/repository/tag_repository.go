package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type tagRepository struct {
	sharedrepo.GenericRepository[entity.Tag]
	db     *core.Database
	logger core.Logger
}

func (r *tagRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func NewTagRepository(db *core.Database, logger core.Logger) repository.TagRepository {
	base := sharedrepo.NewBaseRepository[entity.Tag](db, logger)
	return &tagRepository{
		GenericRepository: base,
		db:                db,
		logger:            logger,
	}
}

func (r *tagRepository) UpsertTranslation(ctx context.Context, trans *entity.TagTranslation) error {
	if err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "tag_id"}, {Name: "language"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"name":        trans.Name,
			"description": trans.Description,
			"updated_at":  gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(trans).Error; err != nil {
		r.logger.Error("Failed to upsert tag translation", core.Error(err))
		return apperrors.InternalError("errors.databaseError", err)
	}
	return nil
}

type sectorRepository struct {
	sharedrepo.GenericRepository[entity.Sector]
	db     *core.Database
	logger core.Logger
}

func (r *sectorRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func NewSectorRepository(db *core.Database, logger core.Logger) repository.SectorRepository {
	base := sharedrepo.NewBaseRepository[entity.Sector](db, logger)
	return &sectorRepository{
		GenericRepository: base,
		db:                db,
		logger:            logger,
	}
}

func (r *sectorRepository) UpsertTranslation(ctx context.Context, trans *entity.SectorTranslation) error {
	if err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "sector_id"}, {Name: "language"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"name":        trans.Name,
			"description": trans.Description,
			"updated_at":  gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(trans).Error; err != nil {
		r.logger.Error("Failed to upsert sector translation", core.Error(err))
		return apperrors.InternalError("errors.databaseError", err)
	}
	return nil
}
