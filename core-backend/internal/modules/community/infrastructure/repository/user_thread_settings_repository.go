package repository

import (
	"context"
	"time"

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

type userThreadSettingsRepository struct {
	db     *core.Database
	logger core.Logger
}

func NewUserThreadSettingsRepository(db *core.Database, logger core.Logger) communityrepo.UserThreadSettingsRepository {
	return &userThreadSettingsRepository{db: db, logger: logger}
}

func (r *userThreadSettingsRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *userThreadSettingsRepository) Get(ctx context.Context, accountID, threadID uuid.UUID) (*entity.UserThreadSettings, error) {
	var settings entity.UserThreadSettings
	if err := r.getDB(ctx).Where("account_id = ? AND thread_id = ?", accountID, threadID).First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, communityerror.ErrThreadSettingNotFound
		}
		r.logger.Error("Failed to get thread settings", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &settings, nil
}

func (r *userThreadSettingsRepository) ListFollowed(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.UserThreadSettings, error) {
	var settings []*entity.UserThreadSettings
	db := r.getDB(ctx).
		Model(&entity.UserThreadSettings{}).
		Where("user_thread_settings.account_id = ? AND user_thread_settings.is_following = ?", accountID, true).
		Joins("JOIN discussion_threads ON discussion_threads.id = user_thread_settings.thread_id").
		Where("discussion_threads.deleted_at IS NULL").
		Preload("Thread")

	if q.Search != "" {
		search := "%" + q.Search + "%"
		db = db.Where(
			"discussion_threads.title ILIKE ? OR discussion_threads.slug ILIKE ? OR discussion_threads.description ILIKE ?",
			search, search, search,
		)
	}

	db = applyPaginationAndSorting(db, q, "user_thread_settings.updated_at desc")
	if err := db.Find(&settings).Error; err != nil {
		r.logger.Error("Failed to list followed threads", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	return settings, nil
}

func (r *userThreadSettingsRepository) UpsertFollow(ctx context.Context, accountID, threadID uuid.UUID, following bool) error {
	settings := &entity.UserThreadSettings{
		AccountID:   accountID,
		ThreadID:    threadID,
		IsFollowing: following,
	}
	if err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "account_id"}, {Name: "thread_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"is_following": following,
			"updated_at":   gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(settings).Error; err != nil {
		r.logger.Error("Failed to upsert thread follow", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *userThreadSettingsRepository) SetMuted(ctx context.Context, accountID, threadID uuid.UUID, muted bool) error {
	settings := &entity.UserThreadSettings{
		AccountID:   accountID,
		ThreadID:    threadID,
		IsMuted:     muted,
		IsFollowing: true,
	}
	if err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "account_id"}, {Name: "thread_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"is_muted":   muted,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(settings).Error; err != nil {
		r.logger.Error("Failed to update thread mute state", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *userThreadSettingsRepository) UpdateLastRead(ctx context.Context, accountID, threadID uuid.UUID, at time.Time) error {
	settings := &entity.UserThreadSettings{
		AccountID:   accountID,
		ThreadID:    threadID,
		IsFollowing: true,
		LastReadAt:  &at,
	}
	if err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "account_id"}, {Name: "thread_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"last_read_at": at,
			"updated_at":   gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(settings).Error; err != nil {
		r.logger.Error("Failed to update thread last read", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *userThreadSettingsRepository) Delete(ctx context.Context, accountID, threadID uuid.UUID) error {
	result := r.getDB(ctx).Where("account_id = ? AND thread_id = ?", accountID, threadID).Delete(&entity.UserThreadSettings{})
	if result.Error != nil {
		r.logger.Error("Failed to delete thread settings", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return communityerror.ErrThreadSettingNotFound
	}
	return nil
}
