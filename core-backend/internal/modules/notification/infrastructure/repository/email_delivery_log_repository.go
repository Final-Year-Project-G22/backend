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

type emailDeliveryLogRepository struct {
	sharedrepo.GenericRepository[entity.EmailDeliveryLog]
	db     *core.Database
	logger core.Logger
}

func NewEmailDeliveryLogRepository(db *core.Database, logger core.Logger) notifrepo.EmailDeliveryLogRepository {
	base := sharedrepo.NewBaseRepository[entity.EmailDeliveryLog](db, logger)
	return &emailDeliveryLogRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *emailDeliveryLogRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *emailDeliveryLogRepository) GetByProviderMessageID(ctx context.Context, providerMessageID string) (*entity.EmailDeliveryLog, error) {
	var log entity.EmailDeliveryLog
	if err := r.getDB(ctx).Where("provider_message_id = ?", providerMessageID).First(&log).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, notiferror.ErrDeliveryLogNotFound
		}
		r.logger.Error("Failed to get delivery log by provider message ID", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &log, nil
}

func (r *emailDeliveryLogRepository) UpdateDeliveryEvent(ctx context.Context, id uuid.UUID, eventType string, occurredAt time.Time, metadata map[string]interface{}) error {
	updates := map[string]interface{}{
		"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
	}

	switch eventType {
	case "delivered":
		updates["delivered_at"] = occurredAt
		updates["delivery_status"] = entity.DeliveryStatusDelivered
	case "opened":
		updates["opened_at"] = occurredAt
	case "clicked":
		updates["clicked_at"] = occurredAt
	case "bounced":
		updates["delivery_status"] = entity.DeliveryStatusBounced
		if reason, ok := metadata["bounceReason"].(string); ok {
			updates["bounce_reason"] = reason
		}
	case "complained":
		updates["complaint"] = true
		updates["delivery_status"] = entity.DeliveryStatusBounced
	}

	result := r.getDB(ctx).Model(&entity.EmailDeliveryLog{}).
		Where("id = ?", id).
		Updates(updates)
	if result.Error != nil {
		r.logger.Error("Failed to update delivery event", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return notiferror.ErrDeliveryLogNotFound
	}
	return nil
}

func (r *emailDeliveryLogRepository) GetByNotificationHistoryID(ctx context.Context, historyID uuid.UUID) (*entity.EmailDeliveryLog, error) {
	var log entity.EmailDeliveryLog
	if err := r.getDB(ctx).Where("notification_history_id = ?", historyID).First(&log).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, notiferror.ErrDeliveryLogNotFound
		}
		r.logger.Error("Failed to get delivery log by history ID", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &log, nil
}
