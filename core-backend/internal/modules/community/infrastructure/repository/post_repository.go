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
