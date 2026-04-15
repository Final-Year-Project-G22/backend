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

type userCategorySettingsRepository struct {
	db     *core.Database
	logger core.Logger
}

func NewUserCategorySettingsRepository(db *core.Database, logger core.Logger) communityrepo.UserCategorySettingsRepository {
	return &userCategorySettingsRepository{db: db, logger: logger}
}

func (r *userCategorySettingsRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *userCategorySettingsRepository) Get(ctx context.Context, accountID, categoryID uuid.UUID) (*entity.UserCategorySettings, error) {
	var settings entity.UserCategorySettings
	if err := r.getDB(ctx).Where("account_id = ? AND category_id = ?", accountID, categoryID).First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, communityerror.ErrCategorySettingNotFound
		}
		r.logger.Error("Failed to get category settings", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &settings, nil
}

func (r *userCategorySettingsRepository) ListFollowed(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.UserCategorySettings, error) {
	var settings []*entity.UserCategorySettings
	db := r.getDB(ctx).
		Model(&entity.UserCategorySettings{}).
		Where("user_category_settings.account_id = ? AND user_category_settings.is_following = ?", accountID, true).
		Joins("JOIN community_categories ON community_categories.id = user_category_settings.category_id").
		Where("community_categories.deleted_at IS NULL").
		Preload("Category")

	if q.Search != "" {
		search := "%" + q.Search + "%"
		db = db.Where(
			"community_categories.name ILIKE ? OR community_categories.slug ILIKE ?",
			search, search,
		)
	}

	db = applyPaginationAndSorting(db, q, "user_category_settings.updated_at desc")
	if err := db.Find(&settings).Error; err != nil {
		r.logger.Error("Failed to list followed categories", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	return settings, nil
}

func (r *userCategorySettingsRepository) UpsertFollow(ctx context.Context, accountID, categoryID uuid.UUID, following bool) error {
	settings := &entity.UserCategorySettings{
		AccountID:   accountID,
		CategoryID:  categoryID,
		IsFollowing: following,
	}
	if err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "account_id"}, {Name: "category_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"is_following": following,
			"updated_at":   gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(settings).Error; err != nil {
		r.logger.Error("Failed to upsert category follow", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *userCategorySettingsRepository) SetMuted(ctx context.Context, accountID, categoryID uuid.UUID, muted bool) error {
	settings := &entity.UserCategorySettings{
		AccountID:   accountID,
		CategoryID:  categoryID,
		IsMuted:     muted,
		IsFollowing: true,
	}
	if err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "account_id"}, {Name: "category_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"is_muted":   muted,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(settings).Error; err != nil {
		r.logger.Error("Failed to update category mute state", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *userCategorySettingsRepository) UpdateLastRead(ctx context.Context, accountID, categoryID uuid.UUID, at time.Time) error {
	settings := &entity.UserCategorySettings{
		AccountID:   accountID,
		CategoryID:  categoryID,
		IsFollowing: true,
		LastReadAt:  &at,
	}
	if err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "account_id"}, {Name: "category_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"last_read_at": at,
			"updated_at":   gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(settings).Error; err != nil {
		r.logger.Error("Failed to update category last read", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *userCategorySettingsRepository) Delete(ctx context.Context, accountID, categoryID uuid.UUID) error {
	result := r.getDB(ctx).Where("account_id = ? AND category_id = ?", accountID, categoryID).Delete(&entity.UserCategorySettings{})
	if result.Error != nil {
		r.logger.Error("Failed to delete category settings", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return communityerror.ErrCategorySettingNotFound
	}
	return nil
}
