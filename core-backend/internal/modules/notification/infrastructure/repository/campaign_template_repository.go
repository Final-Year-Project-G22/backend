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

type campaignTemplateRepository struct {
	sharedrepo.GenericRepository[entity.CampaignTemplate]
	db     *core.Database
	logger core.Logger
}

func NewCampaignTemplateRepository(db *core.Database, logger core.Logger) notifrepo.CampaignTemplateRepository {
	base := sharedrepo.NewBaseRepository[entity.CampaignTemplate](db, logger)
	return &campaignTemplateRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *campaignTemplateRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *campaignTemplateRepository) GetTranslation(ctx context.Context, templateID uuid.UUID, language string) (*entity.CampaignTemplateTranslation, error) {
	var translation entity.CampaignTemplateTranslation
	if err := r.getDB(ctx).Where("campaign_template_id = ? AND language = ?", templateID, language).First(&translation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, notiferror.ErrCampaignTranslationNotFound
		}
		r.logger.Error("Failed to get campaign template translation", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &translation, nil
}

func (r *campaignTemplateRepository) GetTranslations(ctx context.Context, templateID uuid.UUID) ([]*entity.CampaignTemplateTranslation, error) {
	var translations []*entity.CampaignTemplateTranslation
	if err := r.getDB(ctx).Where("campaign_template_id = ?", templateID).Order("language asc").Find(&translations).Error; err != nil {
		r.logger.Error("Failed to get campaign template translations", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return translations, nil
}

func (r *campaignTemplateRepository) UpsertTranslation(ctx context.Context, translation *entity.CampaignTemplateTranslation) error {
	if err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "campaign_template_id"}, {Name: "language"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"content":    translation.Content,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(translation).Error; err != nil {
		r.logger.Error("Failed to upsert campaign template translation", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *campaignTemplateRepository) DeleteTranslation(ctx context.Context, templateID uuid.UUID, language string) error {
	result := r.getDB(ctx).Where("campaign_template_id = ? AND language = ?", templateID, language).Delete(&entity.CampaignTemplateTranslation{})
	if result.Error != nil {
		r.logger.Error("Failed to delete campaign template translation", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return notiferror.ErrCampaignTranslationNotFound
	}
	return nil
}
