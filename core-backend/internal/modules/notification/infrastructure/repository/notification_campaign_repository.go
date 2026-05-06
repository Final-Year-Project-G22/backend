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

type notificationCampaignRepository struct {
	sharedrepo.GenericRepository[entity.NotificationCampaign]
	db     *core.Database
	logger core.Logger
}

func NewNotificationCampaignRepository(db *core.Database, logger core.Logger) notifrepo.NotificationCampaignRepository {
	base := sharedrepo.NewBaseRepository[entity.NotificationCampaign](db, logger)
	return &notificationCampaignRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *notificationCampaignRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *notificationCampaignRepository) ListByStatus(ctx context.Context, status entity.CampaignStatus, q query.QueryOptions) ([]*entity.NotificationCampaign, int64, error) {
	var total int64
	baseDB := r.getDB(ctx).Where("status = ?", status)

	if err := baseDB.Model(&entity.NotificationCampaign{}).Count(&total).Error; err != nil {
		r.logger.Error("Failed to count campaigns by status", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}

	var campaigns []*entity.NotificationCampaign
	db := applyPaginationAndSorting(baseDB, q, "created_at desc")
	if err := db.Find(&campaigns).Error; err != nil {
		r.logger.Error("Failed to list campaigns by status", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}
	return campaigns, total, nil
}

func (r *notificationCampaignRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.CampaignStatus) error {
	result := r.getDB(ctx).Model(&entity.NotificationCampaign{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		r.logger.Error("Failed to update campaign status", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return notiferror.ErrCampaignNotFound
	}
	return nil
}

func (r *notificationCampaignRepository) ListScheduled(ctx context.Context) ([]*entity.NotificationCampaign, error) {
	var campaigns []*entity.NotificationCampaign
	if err := r.getDB(ctx).
		Where("status = ? AND scheduled_for <= ?", entity.CampaignStatusScheduled, time.Now()).
		Order("scheduled_for asc").
		Find(&campaigns).Error; err != nil {
		r.logger.Error("Failed to list scheduled campaigns", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return campaigns, nil
}

func (r *notificationCampaignRepository) GetByCreator(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.NotificationCampaign, error) {
	var campaigns []*entity.NotificationCampaign
	db := r.getDB(ctx).Where("created_by = ?", accountID)
	db = applyPaginationAndSorting(db, q, "created_at desc")
	if err := db.Find(&campaigns).Error; err != nil {
		r.logger.Error("Failed to get campaigns by creator", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return campaigns, nil
}
