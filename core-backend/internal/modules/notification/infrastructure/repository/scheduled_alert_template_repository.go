package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notiferror "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/error"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"gorm.io/gorm"
)

type scheduledAlertTemplateRepository struct {
	sharedrepo.GenericRepository[entity.ScheduledAlertTemplate]
	db     *core.Database
	logger core.Logger
}

func NewScheduledAlertTemplateRepository(db *core.Database, logger core.Logger) notifrepo.ScheduledAlertTemplateRepository {
	base := sharedrepo.NewBaseRepository[entity.ScheduledAlertTemplate](db, logger)
	return &scheduledAlertTemplateRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *scheduledAlertTemplateRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *scheduledAlertTemplateRepository) ListActive(ctx context.Context) ([]*entity.ScheduledAlertTemplate, error) {
	var templates []*entity.ScheduledAlertTemplate
	if err := r.getDB(ctx).Where("is_active = ?", true).Order("slug asc").Find(&templates).Error; err != nil {
		r.logger.Error("Failed to list active scheduled alert templates", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return templates, nil
}

func (r *scheduledAlertTemplateRepository) GetBySlug(ctx context.Context, slug string) (*entity.ScheduledAlertTemplate, error) {
	var tmpl entity.ScheduledAlertTemplate
	if err := r.getDB(ctx).Where("slug = ?", slug).First(&tmpl).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, notiferror.ErrScheduledAlertTemplateNotFound
		}
		r.logger.Error("Failed to get scheduled alert template by slug", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &tmpl, nil
}
