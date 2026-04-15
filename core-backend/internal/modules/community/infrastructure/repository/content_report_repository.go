package repository

import (
	"context"
	"time"

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

type contentReportRepository struct {
	sharedrepo.GenericRepository[entity.ContentReport]
	db     *core.Database
	logger core.Logger
}

func NewContentReportRepository(db *core.Database, logger core.Logger) communityrepo.ContentReportRepository {
	base := sharedrepo.NewBaseRepository[entity.ContentReport](db, logger)
	return &contentReportRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *contentReportRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *contentReportRepository) ListByTarget(ctx context.Context, targetType entity.TargetType, targetID uuid.UUID, q query.QueryOptions) ([]*entity.ContentReport, error) {
	var reports []*entity.ContentReport
	db := r.getDB(ctx).Where("target_type = ? AND target_id = ?", targetType, targetID)
	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	db = applyPaginationAndSorting(db, q, "created_at desc")
	if err := db.Find(&reports).Error; err != nil {
		r.logger.Error("Failed to list reports by target", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return reports, nil
}

func (r *contentReportRepository) ListByStatus(ctx context.Context, status entity.ReportStatus, q query.QueryOptions) ([]*entity.ContentReport, error) {
	var reports []*entity.ContentReport
	db := r.getDB(ctx).Where("status = ?", status)
	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	db = applyPaginationAndSorting(db, q, "created_at desc")
	if err := db.Find(&reports).Error; err != nil {
		r.logger.Error("Failed to list reports by status", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return reports, nil
}

func (r *contentReportRepository) UpdateStatus(ctx context.Context, reportID uuid.UUID, status entity.ReportStatus, adminNote *string, resolvedBy *uuid.UUID, resolvedAt *time.Time) error {
	updates := map[string]interface{}{
		"status":     status,
		"admin_note": adminNote,
		"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
	}
	if resolvedBy != nil {
		updates["resolved_by_account_id"] = *resolvedBy
	}
	if resolvedAt != nil {
		updates["resolved_at"] = *resolvedAt
	}
	result := r.getDB(ctx).Model(&entity.ContentReport{}).Where("id = ?", reportID).Updates(updates)
	if result.Error != nil {
		r.logger.Error("Failed to update report status", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return communityerror.ErrReportNotFound
	}
	return nil
}
