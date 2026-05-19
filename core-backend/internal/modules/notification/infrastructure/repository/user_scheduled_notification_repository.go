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

type userScheduledNotificationRepository struct {
	sharedrepo.GenericRepository[entity.UserScheduledNotification]
	db     *core.Database
	logger core.Logger
}

func NewUserScheduledNotificationRepository(db *core.Database, logger core.Logger) notifrepo.UserScheduledNotificationRepository {
	base := sharedrepo.NewBaseRepository[entity.UserScheduledNotification](db, logger)
	return &userScheduledNotificationRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *userScheduledNotificationRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *userScheduledNotificationRepository) FetchDue(ctx context.Context, limit int) ([]*entity.UserScheduledNotification, error) {
	var items []*entity.UserScheduledNotification
	if err := r.getDB(ctx).
		Where("status = ? AND scheduled_for <= ?", entity.ScheduleStatusPending, time.Now().UTC()).
		Order("scheduled_for asc").
		Limit(limit).
		Find(&items).Error; err != nil {
		r.logger.Error("Failed to fetch due scheduled notifications", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return items, nil
}

func (r *userScheduledNotificationRepository) CountPendingByAccount(ctx context.Context, accountID uuid.UUID) (int64, error) {
	var count int64
	if err := r.getDB(ctx).Model(&entity.UserScheduledNotification{}).
		Where("account_id = ? AND status = ?", accountID, entity.ScheduleStatusPending).
		Count(&count).Error; err != nil {
		r.logger.Error("Failed to count pending scheduled notifications", core.Error(err))
		return 0, errors.InternalError("errors.databaseError", err)
	}
	return count, nil
}

func (r *userScheduledNotificationRepository) ListByAccount(ctx context.Context, accountID uuid.UUID) ([]*entity.UserScheduledNotification, error) {
	var items []*entity.UserScheduledNotification
	if err := r.getDB(ctx).
		Where("account_id = ?", accountID).
		Order("scheduled_for asc").
		Find(&items).Error; err != nil {
		r.logger.Error("Failed to list scheduled notifications by account", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return items, nil
}

func (r *userScheduledNotificationRepository) MarkSent(ctx context.Context, id uuid.UUID) error {
	result := r.getDB(ctx).Model(&entity.UserScheduledNotification{}).
		Where("id = ? AND status = ?", id, entity.ScheduleStatusPending).
		Updates(map[string]interface{}{
			"status":     entity.ScheduleStatusSent,
			"sent_at":    time.Now().UTC(),
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		r.logger.Error("Failed to mark scheduled notification as sent", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return notiferror.ErrScheduledAlertNotFound
	}
	return nil
}

func (r *userScheduledNotificationRepository) CancelByID(ctx context.Context, id uuid.UUID) error {
	result := r.getDB(ctx).Model(&entity.UserScheduledNotification{}).
		Where("id = ? AND status = ?", id, entity.ScheduleStatusPending).
		Updates(map[string]interface{}{
			"status":       entity.ScheduleStatusCancelled,
			"cancelled_at": time.Now().UTC(),
			"updated_at":   gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		r.logger.Error("Failed to cancel scheduled notification", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return notiferror.ErrScheduledAlertNotFound
	}
	return nil
}
