package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notiferror "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/error"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userDeviceRepository struct {
	sharedrepo.GenericRepository[entity.UserDevice]
	db     *core.Database
	logger core.Logger
}

func NewUserDeviceRepository(db *core.Database, logger core.Logger) notifrepo.UserDeviceRepository {
	base := sharedrepo.NewBaseRepository[entity.UserDevice](db, logger)
	return &userDeviceRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *userDeviceRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *userDeviceRepository) ListByAccount(ctx context.Context, accountID uuid.UUID) ([]*entity.UserDevice, error) {
	var devices []*entity.UserDevice
	if err := r.getDB(ctx).Where("account_id = ? AND is_active = ?", accountID, true).Find(&devices).Error; err != nil {
		r.logger.Error("Failed to list user devices", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return devices, nil
}

func (r *userDeviceRepository) GetByDeviceToken(ctx context.Context, deviceToken string) (*entity.UserDevice, error) {
	var device entity.UserDevice
	if err := r.getDB(ctx).Where("device_token = ?", deviceToken).First(&device).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, notiferror.ErrDeviceNotFound
		}
		r.logger.Error("Failed to get device by token", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &device, nil
}

func (r *userDeviceRepository) DeactivateByAccount(ctx context.Context, accountID uuid.UUID) error {
	result := r.getDB(ctx).Model(&entity.UserDevice{}).
		Where("account_id = ?", accountID).
		Update("is_active", false)
	if result.Error != nil {
		r.logger.Error("Failed to deactivate devices by account", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	return nil
}

func (r *userDeviceRepository) UpdatePushToken(ctx context.Context, id uuid.UUID, pushToken string) error {
	result := r.getDB(ctx).Model(&entity.UserDevice{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"push_token": pushToken,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		r.logger.Error("Failed to update push token", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return notiferror.ErrDeviceNotFound
	}
	return nil
}
