package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	communityrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type discussionPostRepository struct {
	sharedrepo.GenericRepository[entity.DiscussionPost]
	db     *core.Database
	logger core.Logger
}

func NewDiscussionPostRepository(db *core.Database, logger core.Logger) communityrepo.DiscussionPostRepository {
	base := sharedrepo.NewBaseRepository[entity.DiscussionPost](db, logger)
	return &discussionPostRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *discussionPostRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *discussionPostRepository) ListByThread(ctx context.Context, threadID uuid.UUID, q query.QueryOptions) ([]*entity.DiscussionPost, error) {
	var posts []*entity.DiscussionPost
	db := r.getDB(ctx).Where("thread_id = ?", threadID)
	if q.Search != "" {
		search := "%" + q.Search + "%"
		db = db.Where("content ILIKE ?", search)
	}
	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	db = applyPaginationAndSorting(db, q, "created_at asc")
	if err := db.Find(&posts).Error; err != nil {
		r.logger.Error("Failed to list posts by thread", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return posts, nil
}

func (r *discussionPostRepository) ListReplies(ctx context.Context, parentPostID uuid.UUID, q query.QueryOptions) ([]*entity.DiscussionPost, error) {
	var posts []*entity.DiscussionPost
	db := r.getDB(ctx).Where("parent_post_id = ?", parentPostID)
	if q.Search != "" {
		search := "%" + q.Search + "%"
		db = db.Where("content ILIKE ?", search)
	}
	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	db = applyPaginationAndSorting(db, q, "created_at asc")
	if err := db.Find(&posts).Error; err != nil {
		r.logger.Error("Failed to list replies", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return posts, nil
}

func (r *discussionPostRepository) GetSolution(ctx context.Context, threadID uuid.UUID) (*entity.DiscussionPost, error) {
	var post entity.DiscussionPost
	if err := r.getDB(ctx).Where("thread_id = ? AND is_solution = ?", threadID, true).First(&post).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Error("Failed to get solution post", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &post, nil
}

func (r *discussionPostRepository) ClearSolution(ctx context.Context, threadID uuid.UUID) error {
	result := r.getDB(ctx).Model(&entity.DiscussionPost{}).
		Where("thread_id = ? AND is_solution = ?", threadID, true).
		Update("is_solution", false)
	if result.Error != nil {
		r.logger.Error("Failed to clear solution post", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	return nil
}

func (r *discussionPostRepository) IsAuthor(ctx context.Context, postID, accountID uuid.UUID) (bool, error) {
	var count int64
	if err := r.getDB(ctx).Model(&entity.DiscussionPost{}).
		Where("id = ? AND author_account_id = ?", postID, accountID).
		Count(&count).Error; err != nil {
		r.logger.Error("Failed to check post author", core.Error(err))
		return false, errors.InternalError("errors.databaseError", err)
	}
	return count > 0, nil
}

func (r *discussionPostRepository) CountUnreadByThreadIDs(ctx context.Context, accountID uuid.UUID, threadIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	result := make(map[uuid.UUID]int, len(threadIDs))
	for _, id := range threadIDs {
		result[id] = 0
	}
	if len(threadIDs) == 0 {
		return result, nil
	}

	var rows []struct {
		ThreadID    uuid.UUID `gorm:"column:thread_id"`
		UnreadCount int       `gorm:"column:unread_count"`
	}

	if err := r.getDB(ctx).Raw(`
		SELECT
			d.thread_id,
			COUNT(*) as unread_count
		FROM discussion_posts d
		LEFT JOIN user_thread_settings u
			ON u.thread_id = d.thread_id AND u.account_id = ?
		WHERE d.thread_id IN ?
			AND d.author_account_id != ?
			AND (u.last_read_at IS NULL OR d.created_at > u.last_read_at)
		GROUP BY d.thread_id
	`, accountID, threadIDs, accountID).Scan(&rows).Error; err != nil {
		r.logger.Error("Failed to count unread posts", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	for _, row := range rows {
		result[row.ThreadID] = row.UnreadCount
	}
	return result, nil
}

func (r *discussionPostRepository) ListSolutionStatus(ctx context.Context, threadIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	result := make(map[uuid.UUID]bool, len(threadIDs))
	for _, id := range threadIDs {
		result[id] = false
	}
	if len(threadIDs) == 0 {
		return result, nil
	}

	var rows []struct {
		ThreadID uuid.UUID `gorm:"column:thread_id"`
	}
	if err := r.getDB(ctx).Model(&entity.DiscussionPost{}).
		Where("thread_id IN ? AND is_solution = ?", threadIDs, true).
		Pluck("thread_id", &rows).Error; err != nil {
		r.logger.Error("Failed to list solution status", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	for _, row := range rows {
		result[row.ThreadID] = true
	}
	return result, nil
}
