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
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type notificationQueueRepository struct {
	sharedrepo.GenericRepository[entity.NotificationQueue]
	db     *core.Database
	logger core.Logger
}

func NewNotificationQueueRepository(db *core.Database, logger core.Logger) notifrepo.NotificationQueueRepository {
	base := sharedrepo.NewBaseRepository[entity.NotificationQueue](db, logger)
	return &notificationQueueRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *notificationQueueRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *notificationQueueRepository) FetchPending(ctx context.Context, limit int) ([]*entity.NotificationQueue, error) {
	var items []*entity.NotificationQueue
	if err := r.getDB(ctx).
		Where("status = ? AND scheduled_for <= ?", entity.NotificationStatusPending, time.Now()).
		Order("priority desc, scheduled_for asc").
		Limit(limit).
		Find(&items).Error; err != nil {
		r.logger.Error("Failed to fetch pending queue items", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return items, nil
}

func (r *notificationQueueRepository) MarkProcessing(ctx context.Context, id uuid.UUID) error {
	result := r.getDB(ctx).Model(&entity.NotificationQueue{}).
		Where("id = ?", id).
		Update("status", entity.NotificationStatusProcessing)
	if result.Error != nil {
		r.logger.Error("Failed to mark queue item as processing", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return notiferror.ErrQueueItemNotFound
	}
	return nil
}

func (r *notificationQueueRepository) MarkDelivered(ctx context.Context, id uuid.UUID) error {
	result := r.getDB(ctx).Model(&entity.NotificationQueue{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     entity.NotificationStatusDelivered,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		r.logger.Error("Failed to mark queue item as delivered", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return notiferror.ErrQueueItemNotFound
	}
	return nil
}

func (r *notificationQueueRepository) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	result := r.getDB(ctx).Model(&entity.NotificationQueue{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        entity.NotificationStatusFailed,
			"error_message": errMsg,
			"updated_at":    gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		r.logger.Error("Failed to mark queue item as failed", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return notiferror.ErrQueueItemNotFound
	}
	return nil
}

func (r *notificationQueueRepository) IncrementRetry(ctx context.Context, id uuid.UUID, nextScheduledFor time.Time) error {
	result := r.getDB(ctx).Model(&entity.NotificationQueue{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"retry_count":   gorm.Expr("retry_count + 1"),
			"scheduled_for": nextScheduledFor,
			"status":        entity.NotificationStatusPending,
			"error_message": gorm.Expr("NULL"),
			"updated_at":    gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		r.logger.Error("Failed to increment retry", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return notiferror.ErrQueueItemNotFound
	}
	return nil
}

func (r *notificationQueueRepository) CancelByAccount(ctx context.Context, accountID uuid.UUID) error {
	result := r.getDB(ctx).Model(&entity.NotificationQueue{}).
		Where("account_id = ? AND status = ?", accountID, entity.NotificationStatusPending).
		Updates(map[string]interface{}{
			"status":     entity.NotificationStatusCancelled,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		r.logger.Error("Failed to cancel queue items by account", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	return nil
}

func (r *notificationQueueRepository) CancelByCampaign(ctx context.Context, campaignID uuid.UUID) error {
	result := r.getDB(ctx).Model(&entity.NotificationQueue{}).
		Where("campaign_id = ? AND status = ?", campaignID, entity.NotificationStatusPending).
		Updates(map[string]interface{}{
			"status":     entity.NotificationStatusCancelled,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		r.logger.Error("Failed to cancel queue items by campaign", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	return nil
}

func (r *notificationQueueRepository) CountByStatus(ctx context.Context, status entity.NotificationStatus) (int64, error) {
	var count int64
	if err := r.getDB(ctx).Model(&entity.NotificationQueue{}).Where("status = ?", status).Count(&count).Error; err != nil {
		r.logger.Error("Failed to count queue items by status", core.Error(err))
		return 0, errors.InternalError("errors.databaseError", err)
	}
	return count, nil
}
