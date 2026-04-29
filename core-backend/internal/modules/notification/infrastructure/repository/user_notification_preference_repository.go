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
	"gorm.io/gorm/clause"
)

type userNotificationPreferenceRepository struct {
	sharedrepo.GenericRepository[entity.UserNotificationPreference]
	db     *core.Database
	logger core.Logger
}

func NewUserNotificationPreferenceRepository(db *core.Database, logger core.Logger) notifrepo.UserNotificationPreferenceRepository {
	base := sharedrepo.NewBaseRepository[entity.UserNotificationPreference](db, logger)
	return &userNotificationPreferenceRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *userNotificationPreferenceRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *userNotificationPreferenceRepository) GetByAccountAndTypeAndChannel(ctx context.Context, accountID uuid.UUID, notificationType entity.NotificationType, channel entity.Channel) (*entity.UserNotificationPreference, error) {
	var pref entity.UserNotificationPreference
	if err := r.getDB(ctx).
		Where("account_id = ? AND notification_type = ? AND channel = ?", accountID, notificationType, channel).
		First(&pref).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, notiferror.ErrPreferenceNotFound
		}
		r.logger.Error("Failed to get preference", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &pref, nil
}

func (r *userNotificationPreferenceRepository) ListByAccount(ctx context.Context, accountID uuid.UUID) ([]*entity.UserNotificationPreference, error) {
	var prefs []*entity.UserNotificationPreference
	if err := r.getDB(ctx).Where("account_id = ?", accountID).Find(&prefs).Error; err != nil {
		r.logger.Error("Failed to list preferences", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return prefs, nil
}

func (r *userNotificationPreferenceRepository) Upsert(ctx context.Context, pref *entity.UserNotificationPreference) error {
	if err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "account_id"},
			{Name: "notification_type"},
			{Name: "channel"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"is_enabled":        pref.IsEnabled,
			"quiet_hours_start": pref.QuietHoursStart,
			"quiet_hours_end":   pref.QuietHoursEnd,
			"updated_at":        gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(pref).Error; err != nil {
		r.logger.Error("Failed to upsert preference", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}
