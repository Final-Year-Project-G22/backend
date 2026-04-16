package infrastructure

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/port"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DLQController struct {
	db     *core.Database
	logger core.Logger
}

func NewDLQController(db *core.Database, logger core.Logger) *DLQController {
	return &DLQController{
		db:     db,
		logger: logger,
	}
}

var _ port.DLQController = (*DLQController)(nil)

func (c *DLQController) ListDeadEvents(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*entity.IngestionOutbox, error) {
	var events []*entity.IngestionOutbox
	err := c.db.WithContext(ctx).
		Where("account_id = ? AND status = ?", accountID, entity.OutboxStatusDead).
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	if err != nil {
		c.logger.Error("Failed to list dead events", core.Error(err))
		return nil, err
	}
	return events, nil
}

func (c *DLQController) GetDeadEvent(ctx context.Context, eventID uuid.UUID) (*entity.IngestionOutbox, error) {
	var event entity.IngestionOutbox
	err := c.db.WithContext(ctx).
		Where("event_id = ? AND status = ?", eventID, entity.OutboxStatusDead).
		First(&event).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		c.logger.Error("Failed to get dead event", core.Error(err))
		return nil, err
	}
	return &event, nil
}

func (c *DLQController) ReDriveEvent(ctx context.Context, eventID uuid.UUID, operatorID uuid.UUID) error {
	return c.db.Transaction(ctx, func(tx *gorm.DB) error {
		var event entity.IngestionOutbox
		err := tx.Where("event_id = ? AND status = ?", eventID, entity.OutboxStatusDead).First(&event).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return gorm.ErrRecordNotFound
			}
			return err
		}

		event.Status = entity.OutboxStatusPending
		event.AttemptCount = 0
		event.LastError = nil
		event.ReplayCount++

		if err := tx.Save(&event).Error; err != nil {
			c.logger.Error("Failed to re-drive event", core.Error(err))
			return err
		}

		c.logger.Info("Event re-driven",
			core.String("event_id", eventID.String()),
			core.String("operator_id", operatorID.String()))

		return nil
	})
}

func (c *DLQController) ReDriveBatch(ctx context.Context, eventIDs []uuid.UUID, operatorID uuid.UUID) (int, error) {
	var successCount int
	err := c.db.Transaction(ctx, func(tx *gorm.DB) error {
		for _, eventID := range eventIDs {
			var event entity.IngestionOutbox
			err := tx.Where("event_id = ? AND status = ?", eventID, entity.OutboxStatusDead).First(&event).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					continue
				}
				return err
			}

			event.Status = entity.OutboxStatusPending
			event.AttemptCount = 0
			event.LastError = nil
			event.ReplayCount++

			if err := tx.Save(&event).Error; err != nil {
				c.logger.Error("Failed to re-drive event in batch", core.Error(err))
				continue
			}
			successCount++
		}

		c.logger.Info("Batch re-drive completed",
			core.Int("success_count", successCount),
			core.Int("total_requested", len(eventIDs)),
			core.String("operator_id", operatorID.String()))

		return nil
	})
	if err != nil {
		return successCount, err
	}
	return successCount, nil
}

func (c *DLQController) GetReDriveHistory(ctx context.Context, eventID uuid.UUID) ([]port.DLQReDriveAudit, error) {
	return []port.DLQReDriveAudit{}, nil
}
