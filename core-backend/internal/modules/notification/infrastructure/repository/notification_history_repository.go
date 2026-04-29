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

type notificationHistoryRepository struct {
	sharedrepo.GenericRepository[entity.NotificationHistory]
	db     *core.Database
	logger core.Logger
}

func NewNotificationHistoryRepository(db *core.Database, logger core.Logger) notifrepo.NotificationHistoryRepository {
	base := sharedrepo.NewBaseRepository[entity.NotificationHistory](db, logger)
	return &notificationHistoryRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *notificationHistoryRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *notificationHistoryRepository) ListByAccount(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.NotificationHistory, error) {
	var history []*entity.NotificationHistory
	db := r.getDB(ctx).Where("account_id = ?", accountID)
	db = applyPaginationAndSorting(db, q, "sent_at desc")
	if err := db.Find(&history).Error; err != nil {
		r.logger.Error("Failed to list notification history", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return history, nil
}

func (r *notificationHistoryRepository) UpdateDeliveryStatus(ctx context.Context, id uuid.UUID, status entity.DeliveryStatus, deliveredAt *time.Time) error {
	updates := map[string]interface{}{
		"delivery_status": status,
		"updated_at":      gorm.Expr("CURRENT_TIMESTAMP"),
	}
	if deliveredAt != nil {
		updates["delivered_at"] = *deliveredAt
	}
	result := r.getDB(ctx).Model(&entity.NotificationHistory{}).
		Where("id = ?", id).
		Updates(updates)
	if result.Error != nil {
		r.logger.Error("Failed to update delivery status", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return notiferror.ErrHistoryNotFound
	}
	return nil
}

func (r *notificationHistoryRepository) MarkRead(ctx context.Context, id uuid.UUID) error {
	result := r.getDB(ctx).Model(&entity.NotificationHistory{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"read_at":    time.Now(),
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		r.logger.Error("Failed to mark history as read", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return notiferror.ErrHistoryNotFound
	}
	return nil
}

func (r *notificationHistoryRepository) MarkClicked(ctx context.Context, id uuid.UUID) error {
	result := r.getDB(ctx).Model(&entity.NotificationHistory{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"clicked_at": time.Now(),
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		r.logger.Error("Failed to mark history as clicked", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return notiferror.ErrHistoryNotFound
	}
	return nil
}
