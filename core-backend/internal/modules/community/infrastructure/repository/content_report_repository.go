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

func (r *contentReportRepository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	db := r.getDB(ctx).Model(&entity.ContentReport{})
	if err := db.Count(&count).Error; err != nil {
		r.logger.Error("Failed to count all reports", core.Error(err))
		return 0, errors.InternalError("errors.databaseError", err)
	}
	return count, nil
}

func (r *contentReportRepository) CountByThreadID(ctx context.Context, threadID uuid.UUID) (int64, error) {
	var count int64
	db := r.getDB(ctx).Model(&entity.ContentReport{}).Where("thread_id = ?", threadID)
	if err := db.Count(&count).Error; err != nil {
		r.logger.Error("Failed to count reports by thread ID", core.Error(err))
		return 0, errors.InternalError("errors.databaseError", err)
	}
	return count, nil
}

func (r *contentReportRepository) CountByPostID(ctx context.Context, postID uuid.UUID) (int64, error) {
	var count int64
	db := r.getDB(ctx).Model(&entity.ContentReport{}).Where("post_id = ?", postID)
	if err := db.Count(&count).Error; err != nil {
		r.logger.Error("Failed to count reports by post ID", core.Error(err))
		return 0, errors.InternalError("errors.databaseError", err)
	}
	return count, nil
}

func (r *contentReportRepository) CountByReportedAccountID(ctx context.Context, accountID uuid.UUID) (int64, error) {
	var count int64
	db := r.getDB(ctx).Model(&entity.ContentReport{}).Where("reported_account_id = ?", accountID)
	if err := db.Count(&count).Error; err != nil {
		r.logger.Error("Failed to count reports by reported account ID", core.Error(err))
		return 0, errors.InternalError("errors.databaseError", err)
	}
	return count, nil
}

func (r *contentReportRepository) ListByThreadID(ctx context.Context, threadID uuid.UUID, q query.QueryOptions) ([]*entity.ContentReport, int64, error) {
	var reports []*entity.ContentReport
	db := r.getDB(ctx).Model(&entity.ContentReport{}).Where("thread_id = ?", threadID)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count reports by thread ID", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}

	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	db = applyPaginationAndSorting(db, q, "created_at desc")
	if err := db.Find(&reports).Error; err != nil {
		r.logger.Error("Failed to list reports by thread ID", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}
	return reports, total, nil
}

func (r *contentReportRepository) ListByThreadIDAndStatus(ctx context.Context, threadID uuid.UUID, status entity.ReportStatus, q query.QueryOptions) ([]*entity.ContentReport, int64, error) {
	var reports []*entity.ContentReport
	db := r.getDB(ctx).Model(&entity.ContentReport{}).Where("thread_id = ? AND status = ?", threadID, status)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count reports by thread ID and status", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}

	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	db = applyPaginationAndSorting(db, q, "created_at desc")
	if err := db.Find(&reports).Error; err != nil {
		r.logger.Error("Failed to list reports by thread ID and status", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}
	return reports, total, nil
}

func (r *contentReportRepository) ListByPostID(ctx context.Context, postID uuid.UUID, q query.QueryOptions) ([]*entity.ContentReport, int64, error) {
	var reports []*entity.ContentReport
	db := r.getDB(ctx).Model(&entity.ContentReport{}).Where("post_id = ?", postID)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count reports by post ID", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}

	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	db = applyPaginationAndSorting(db, q, "created_at desc")
	if err := db.Find(&reports).Error; err != nil {
		r.logger.Error("Failed to list reports by post ID", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}
	return reports, total, nil
}

func (r *contentReportRepository) ListByPostIDAndStatus(ctx context.Context, postID uuid.UUID, status entity.ReportStatus, q query.QueryOptions) ([]*entity.ContentReport, int64, error) {
	var reports []*entity.ContentReport
	db := r.getDB(ctx).Model(&entity.ContentReport{}).Where("post_id = ? AND status = ?", postID, status)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count reports by post ID and status", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}

	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	db = applyPaginationAndSorting(db, q, "created_at desc")
	if err := db.Find(&reports).Error; err != nil {
		r.logger.Error("Failed to list reports by post ID and status", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}
	return reports, total, nil
}

func (r *contentReportRepository) ListByReportedAccountID(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.ContentReport, int64, error) {
	var reports []*entity.ContentReport
	db := r.getDB(ctx).Model(&entity.ContentReport{}).Where("reported_account_id = ?", accountID)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count reports by reported account ID", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}

	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	db = applyPaginationAndSorting(db, q, "created_at desc")
	if err := db.Find(&reports).Error; err != nil {
		r.logger.Error("Failed to list reports by reported account ID", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}
	return reports, total, nil
}

func (r *contentReportRepository) ListByReportedAccountIDAndStatus(ctx context.Context, accountID uuid.UUID, status entity.ReportStatus, q query.QueryOptions) ([]*entity.ContentReport, int64, error) {
	var reports []*entity.ContentReport
	db := r.getDB(ctx).Model(&entity.ContentReport{}).Where("reported_account_id = ? AND status = ?", accountID, status)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count reports by reported account ID and status", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}

	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	db = applyPaginationAndSorting(db, q, "created_at desc")
	if err := db.Find(&reports).Error; err != nil {
		r.logger.Error("Failed to list reports by reported account ID and status", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}
	return reports, total, nil
}

func (r *contentReportRepository) ListAllThreadReports(ctx context.Context, q query.QueryOptions) ([]*entity.ContentReport, int64, error) {
	var reports []*entity.ContentReport
	db := r.getDB(ctx).Model(&entity.ContentReport{}).
		Joins("INNER JOIN discussion_threads ON discussion_threads.id = content_reports.thread_id").
		Where("content_reports.thread_id IS NOT NULL AND content_reports.post_id IS NULL AND content_reports.reported_account_id IS NULL AND discussion_threads.deleted_at IS NULL")

	if q.Search != "" {
		search := "%" + q.Search + "%"
		db = db.Where("(discussion_threads.title ILIKE ? OR discussion_threads.slug ILIKE ? OR discussion_threads.description ILIKE ?)", search, search, search)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count all thread reports", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}

	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	db = applyPaginationAndSorting(db, q, "content_reports.created_at desc")
	if err := db.Find(&reports).Error; err != nil {
		r.logger.Error("Failed to list all thread reports", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}
	return reports, total, nil
}

func (r *contentReportRepository) ListAllPostReports(ctx context.Context, q query.QueryOptions) ([]*entity.ContentReport, int64, error) {
	var reports []*entity.ContentReport
	db := r.getDB(ctx).Model(&entity.ContentReport{}).
		Joins("INNER JOIN discussion_posts ON discussion_posts.id = content_reports.post_id").
		Where("content_reports.post_id IS NOT NULL")

	if q.Search != "" {
		search := "%" + q.Search + "%"
		db = db.Where("discussion_posts.content ILIKE ?", search)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count all post reports", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}

	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	db = applyPaginationAndSorting(db, q, "content_reports.created_at desc")
	if err := db.Find(&reports).Error; err != nil {
		r.logger.Error("Failed to list all post reports", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}
	return reports, total, nil
}

func (r *contentReportRepository) ListAllUserReports(ctx context.Context, q query.QueryOptions) ([]*entity.ContentReport, int64, error) {
	var reports []*entity.ContentReport
	db := r.getDB(ctx).Model(&entity.ContentReport{}).
		Joins("INNER JOIN accounts ON accounts.id = content_reports.reported_account_id").
		Joins("INNER JOIN users ON users.id = accounts.user_id").
		Joins("LEFT JOIN thread_blocked_users ON thread_blocked_users.blocked_account_id = content_reports.reported_account_id AND thread_blocked_users.thread_id = content_reports.thread_id").
		Where("content_reports.reported_account_id IS NOT NULL AND thread_blocked_users.blocked_account_id IS NULL")

	if q.Search != "" {
		search := "%" + q.Search + "%"
		db = db.Where("(users.first_name ILIKE ? OR users.last_name ILIKE ? OR accounts.email ILIKE ?)", search, search, search)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count all user reports", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}

	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	db = applyPaginationAndSorting(db, q, "content_reports.created_at desc")
	if err := db.Find(&reports).Error; err != nil {
		r.logger.Error("Failed to list all user reports", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}
	return reports, total, nil
}

func (r *contentReportRepository) ListByThreadStatus(ctx context.Context, status entity.ReportStatus, q query.QueryOptions) ([]*entity.ContentReport, int64, error) {
	var reports []*entity.ContentReport
	db := r.getDB(ctx).Model(&entity.ContentReport{}).
		Joins("INNER JOIN discussion_threads ON discussion_threads.id = content_reports.thread_id").
		Where("content_reports.thread_id IS NOT NULL AND content_reports.post_id IS NULL AND content_reports.reported_account_id IS NULL AND content_reports.status = ? AND discussion_threads.deleted_at IS NULL", status)

	if q.Search != "" {
		search := "%" + q.Search + "%"
		db = db.Where("(discussion_threads.title ILIKE ? OR discussion_threads.slug ILIKE ? OR discussion_threads.description ILIKE ?)", search, search, search)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count reports by thread status", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}

	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	db = applyPaginationAndSorting(db, q, "content_reports.created_at desc")
	if err := db.Find(&reports).Error; err != nil {
		r.logger.Error("Failed to list reports by thread status", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}
	return reports, total, nil
}

func (r *contentReportRepository) ListByPostStatus(ctx context.Context, status entity.ReportStatus, q query.QueryOptions) ([]*entity.ContentReport, int64, error) {
	var reports []*entity.ContentReport
	db := r.getDB(ctx).Model(&entity.ContentReport{}).
		Joins("INNER JOIN discussion_posts ON discussion_posts.id = content_reports.post_id").
		Where("content_reports.post_id IS NOT NULL AND content_reports.status = ?", status)

	if q.Search != "" {
		search := "%" + q.Search + "%"
		db = db.Where("discussion_posts.content ILIKE ?", search)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count reports by post status", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}

	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	db = applyPaginationAndSorting(db, q, "created_at desc")
	if err := db.Find(&reports).Error; err != nil {
		r.logger.Error("Failed to list reports by post status", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}
	return reports, total, nil
}

func (r *contentReportRepository) ListByUserStatus(ctx context.Context, status entity.ReportStatus, q query.QueryOptions) ([]*entity.ContentReport, int64, error) {
	var reports []*entity.ContentReport
	db := r.getDB(ctx).Model(&entity.ContentReport{}).
		Joins("INNER JOIN accounts ON accounts.id = content_reports.reported_account_id").
		Joins("INNER JOIN users ON users.id = accounts.user_id").
		Joins("LEFT JOIN thread_blocked_users ON thread_blocked_users.blocked_account_id = content_reports.reported_account_id AND thread_blocked_users.thread_id = content_reports.thread_id").
		Where("content_reports.reported_account_id IS NOT NULL AND content_reports.status = ? AND thread_blocked_users.blocked_account_id IS NULL", status)

	if q.Search != "" {
		search := "%" + q.Search + "%"
		db = db.Where("(users.first_name ILIKE ? OR users.last_name ILIKE ? OR accounts.email ILIKE ?)", search, search, search)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count reports by user status", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}

	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	db = applyPaginationAndSorting(db, q, "created_at desc")
	if err := db.Find(&reports).Error; err != nil {
		r.logger.Error("Failed to list reports by user status", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}
	return reports, total, nil
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

func (r *contentReportRepository) FindByThreadAndAccount(ctx context.Context, threadID, accountID uuid.UUID) (*entity.ContentReport, error) {
	var report entity.ContentReport
	db := r.getDB(ctx).Where("thread_id = ? AND reported_account_id = ? AND status = ?", threadID, accountID, entity.ReportStatusPending)
	if err := db.First(&report).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Error("Failed to find report by thread and account", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &report, nil
}
