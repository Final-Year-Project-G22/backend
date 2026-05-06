package repository

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notiferror "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/error"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userNotificationInboxRepository struct {
	sharedrepo.GenericRepository[entity.UserNotificationInbox]
	db     *core.Database
	logger core.Logger
}

func NewUserNotificationInboxRepository(db *core.Database, logger core.Logger) notifrepo.UserNotificationInboxRepository {
	base := sharedrepo.NewBaseRepository[entity.UserNotificationInbox](db, logger)
	return &userNotificationInboxRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *userNotificationInboxRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *userNotificationInboxRepository) ListByAccount(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.UserNotificationInbox, int64, error) {
	var total int64
	baseDB := r.getDB(ctx).Where("account_id = ? AND is_archived = ? AND (expires_at IS NULL OR expires_at > ?)", accountID, false, time.Now())

	if err := baseDB.Model(&entity.UserNotificationInbox{}).Count(&total).Error; err != nil {
		r.logger.Error("Failed to count inbox", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}

	var inbox []*entity.UserNotificationInbox
	db := applyPaginationAndSorting(baseDB, q, "created_at desc")
	if err := db.Find(&inbox).Error; err != nil {
		r.logger.Error("Failed to list inbox", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}
	return inbox, total, nil
}

func (r *userNotificationInboxRepository) GetUnreadCount(ctx context.Context, accountID uuid.UUID) (int64, error) {
	var count int64
	if err := r.getDB(ctx).Model(&entity.UserNotificationInbox{}).
		Where("account_id = ? AND is_read = ? AND is_archived = ? AND (expires_at IS NULL OR expires_at > ?)", accountID, false, false, time.Now()).
		Count(&count).Error; err != nil {
		r.logger.Error("Failed to get unread count", core.Error(err))
		return 0, errors.InternalError("errors.databaseError", err)
	}
	return count, nil
}

func (r *userNotificationInboxRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	result := r.getDB(ctx).Model(&entity.UserNotificationInbox{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_read":    true,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		r.logger.Error("Failed to mark inbox entry as read", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return notiferror.ErrInboxEntryNotFound
	}
	return nil
}

func (r *userNotificationInboxRepository) MarkAllAsRead(ctx context.Context, accountID uuid.UUID) error {
	result := r.getDB(ctx).Model(&entity.UserNotificationInbox{}).
		Where("account_id = ? AND is_read = ? AND is_archived = ?", accountID, false, false).
		Updates(map[string]interface{}{
			"is_read":    true,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		r.logger.Error("Failed to mark all inbox as read", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	return nil
}

func (r *userNotificationInboxRepository) Archive(ctx context.Context, id uuid.UUID) error {
	result := r.getDB(ctx).Model(&entity.UserNotificationInbox{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_archived": true,
			"updated_at":  gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		r.logger.Error("Failed to archive inbox entry", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return notiferror.ErrInboxEntryNotFound
	}
	return nil
}

func (r *userNotificationInboxRepository) ExpireOld(ctx context.Context, before time.Time) error {
	result := r.getDB(ctx).Where("expires_at <= ?", before).Delete(&entity.UserNotificationInbox{})
	if result.Error != nil {
		r.logger.Error("Failed to expire old inbox entries", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	return nil
}

func (r *userNotificationInboxRepository) MarkAllReadByCategory(ctx context.Context, accountID uuid.UUID, category entity.NotificationCategory) error {
	result := r.getDB(ctx).Model(&entity.UserNotificationInbox{}).
		Where("account_id = ? AND category = ? AND is_read = ? AND is_archived = ?", accountID, category, false, false).
		Updates(map[string]interface{}{
			"is_read":    true,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		r.logger.Error("Failed to mark category as read", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	return nil
}
