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

type discussionThreadRepository struct {
	sharedrepo.GenericRepository[entity.DiscussionThread]
	db     *core.Database
	logger core.Logger
}

func NewDiscussionThreadRepository(db *core.Database, logger core.Logger) communityrepo.DiscussionThreadRepository {
	base := sharedrepo.NewBaseRepository[entity.DiscussionThread](db, logger)
	return &discussionThreadRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *discussionThreadRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *discussionThreadRepository) GetBySlug(ctx context.Context, slug string, parentThreadID *uuid.UUID) (*entity.DiscussionThread, error) {
	var thread entity.DiscussionThread
	db := r.getDB(ctx).Where("slug = ?", slug)
	if parentThreadID != nil {
		db = db.Where("parent_thread_id = ?", parentThreadID)
	} else {
		db = db.Where("parent_thread_id IS NULL")
	}
	if err := db.First(&thread).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, communityerror.ErrThreadNotFound
		}
		r.logger.Error("Failed to get thread by slug", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &thread, nil
}

func (r *discussionThreadRepository) ListByCategory(ctx context.Context, categoryID uuid.UUID, q query.QueryOptions) ([]*entity.DiscussionThread, error) {
	var threads []*entity.DiscussionThread
	db := r.getDB(ctx).Where("category_id = ?", categoryID)
	if q.Search != "" {
		search := "%" + q.Search + "%"
		db = db.Where("title ILIKE ? OR description ILIKE ? OR slug ILIKE ?", search, search, search)
	}
	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	db = applyPaginationAndSorting(db, q, "last_activity_at desc")
	if err := db.Find(&threads).Error; err != nil {
		r.logger.Error("Failed to list threads by category", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return threads, nil
}

func (r *discussionThreadRepository) Search(ctx context.Context, keyword string, categoryID *uuid.UUID, q query.QueryOptions) ([]*entity.DiscussionThread, error) {
	var threads []*entity.DiscussionThread
	searchTerm := keyword
	if searchTerm == "" {
		searchTerm = q.Search
	}
	db := r.getDB(ctx)
	if categoryID != nil {
		db = db.Where("category_id = ?", *categoryID)
	}
	if searchTerm != "" {
		search := "%" + searchTerm + "%"
		db = db.Where("title ILIKE ? OR description ILIKE ? OR slug ILIKE ?", search, search, search)
	}
	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	db = applyPaginationAndSorting(db, q, "last_activity_at desc")
	if err := db.Find(&threads).Error; err != nil {
		r.logger.Error("Failed to search threads", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return threads, nil
}

func (r *discussionThreadRepository) IncrementViews(ctx context.Context, threadID uuid.UUID) error {
	result := r.getDB(ctx).Model(&entity.DiscussionThread{}).
		Where("id = ?", threadID).
		UpdateColumn("view_count", gorm.Expr("view_count + 1"))
	if result.Error != nil {
		r.logger.Error("Failed to increment thread views", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return communityerror.ErrThreadNotFound
	}
	return nil
}

func (r *discussionThreadRepository) UpdateLastActivity(ctx context.Context, threadID uuid.UUID, at time.Time) error {
	result := r.getDB(ctx).Model(&entity.DiscussionThread{}).
		Where("id = ?", threadID).
		Updates(map[string]interface{}{
			"last_activity_at": at,
			"updated_at":       gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		r.logger.Error("Failed to update thread last activity", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return communityerror.ErrThreadNotFound
	}
	return nil
}

func (r *discussionThreadRepository) UpdateReplyCount(ctx context.Context, threadID uuid.UUID, delta int) error {
	result := r.getDB(ctx).Model(&entity.DiscussionThread{}).
		Where("id = ?", threadID).
		UpdateColumn("reply_count", gorm.Expr("reply_count + ?", delta))
	if result.Error != nil {
		r.logger.Error("Failed to update thread reply count", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return communityerror.ErrThreadNotFound
	}
	return nil
}

func (r *discussionThreadRepository) GetStatus(ctx context.Context, threadID uuid.UUID) (entity.ThreadStatus, error) {
	var thread entity.DiscussionThread
	if err := r.getDB(ctx).Select("status").Where("id = ?", threadID).First(&thread).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", communityerror.ErrThreadNotFound
		}
		r.logger.Error("Failed to get thread status", core.Error(err))
		return "", errors.InternalError("errors.databaseError", err)
	}
	return thread.Status, nil
}

func (r *discussionThreadRepository) IsAuthor(ctx context.Context, threadID, accountID uuid.UUID) (bool, error) {
	var count int64
	if err := r.getDB(ctx).Model(&entity.DiscussionThread{}).
		Where("id = ? AND author_account_id = ?", threadID, accountID).
		Count(&count).Error; err != nil {
		r.logger.Error("Failed to check thread author", core.Error(err))
		return false, errors.InternalError("errors.databaseError", err)
	}
	return count > 0, nil
}
