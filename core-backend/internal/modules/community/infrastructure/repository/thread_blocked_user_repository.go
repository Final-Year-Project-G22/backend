package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	communityerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/error"
	communityrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type threadBlockedUserRepository struct {
	db     *core.Database
	logger core.Logger
}

func NewThreadBlockedUserRepository(db *core.Database, logger core.Logger) communityrepo.ThreadBlockedUserRepository {
	return &threadBlockedUserRepository{db: db, logger: logger}
}

func (r *threadBlockedUserRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *threadBlockedUserRepository) IsBlocked(ctx context.Context, threadID, accountID uuid.UUID) (bool, error) {
	var count int64
	if err := r.getDB(ctx).Model(&entity.ThreadBlockedUser{}).
		Where("thread_id = ? AND blocked_account_id = ?", threadID, accountID).
		Count(&count).Error; err != nil {
		r.logger.Error("Failed to check thread block", core.Error(err))
		return false, errors.InternalError("errors.databaseError", err)
	}
	return count > 0, nil
}

func (r *threadBlockedUserRepository) Block(ctx context.Context, threadID, blockedID, blockedByID uuid.UUID, reason *string) error {
	block := &entity.ThreadBlockedUser{
		ThreadID:           threadID,
		BlockedAccountID:   blockedID,
		BlockedByAccountID: blockedByID,
		Reason:             reason,
	}
	if err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "thread_id"}, {Name: "blocked_account_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"blocked_by_account_id": blockedByID,
			"reason":                reason,
			"updated_at":            gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(block).Error; err != nil {
		r.logger.Error("Failed to block user in thread", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *threadBlockedUserRepository) Unblock(ctx context.Context, threadID, blockedID uuid.UUID) error {
	result := r.getDB(ctx).Where("thread_id = ? AND blocked_account_id = ?", threadID, blockedID).Delete(&entity.ThreadBlockedUser{})
	if result.Error != nil {
		r.logger.Error("Failed to unblock user in thread", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return communityerror.ErrThreadBlockedUserNotFound
	}
	return nil
}

func (r *threadBlockedUserRepository) ListBlocked(ctx context.Context, threadID uuid.UUID, q query.QueryOptions) ([]*entity.ThreadBlockedUser, error) {
	var blocks []*entity.ThreadBlockedUser
	db := r.getDB(ctx).Where("thread_id = ?", threadID)
	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	db = applyPaginationAndSorting(db, q, "created_at desc")
	if err := db.Find(&blocks).Error; err != nil {
		r.logger.Error("Failed to list blocked users", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return blocks, nil
}
