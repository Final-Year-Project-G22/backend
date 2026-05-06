package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notiferror "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/error"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type notificationTemplateRepository struct {
	sharedrepo.GenericRepository[entity.NotificationTemplate]
	db     *core.Database
	logger core.Logger
}

func NewNotificationTemplateRepository(db *core.Database, logger core.Logger) notifrepo.NotificationTemplateRepository {
	base := sharedrepo.NewBaseRepository[entity.NotificationTemplate](db, logger)
	return &notificationTemplateRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *notificationTemplateRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *notificationTemplateRepository) GetByType(ctx context.Context, notificationType entity.NotificationType) (*entity.NotificationTemplate, error) {
	var tmpl entity.NotificationTemplate
	if err := r.getDB(ctx).Where("notification_type = ?", notificationType).First(&tmpl).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, notiferror.ErrTemplateNotFound
		}
		r.logger.Error("Failed to get template by type", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &tmpl, nil
}

func (r *notificationTemplateRepository) ListByTemplateGroup(ctx context.Context, group string, q query.QueryOptions) ([]*entity.NotificationTemplate, error) {
	var templates []*entity.NotificationTemplate
	db := r.getDB(ctx).Where("template_group = ?", group)
	db = applyPaginationAndSorting(db, q, "created_at desc")
	if err := db.Find(&templates).Error; err != nil {
		r.logger.Error("Failed to list templates by group", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return templates, nil
}

func (r *notificationTemplateRepository) GetTranslations(ctx context.Context, templateID uuid.UUID) ([]*entity.NotificationTemplateTranslation, error) {
	var translations []*entity.NotificationTemplateTranslation
	if err := r.getDB(ctx).Where("template_id = ?", templateID).Order("language asc").Find(&translations).Error; err != nil {
		r.logger.Error("Failed to get template translations", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return translations, nil
}

func (r *notificationTemplateRepository) UpsertTranslation(ctx context.Context, translation *entity.NotificationTemplateTranslation) error {
	if err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "template_id"}, {Name: "language"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"subject":    translation.Subject,
			"content":    translation.Content,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(translation).Error; err != nil {
		r.logger.Error("Failed to upsert template translation", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *notificationTemplateRepository) DeleteTranslation(ctx context.Context, templateID uuid.UUID, language string) error {
	result := r.getDB(ctx).Where("template_id = ? AND language = ?", templateID, language).Delete(&entity.NotificationTemplateTranslation{})
	if result.Error != nil {
		r.logger.Error("Failed to delete template translation", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return notiferror.ErrTranslationNotFound
	}
	return nil
}
