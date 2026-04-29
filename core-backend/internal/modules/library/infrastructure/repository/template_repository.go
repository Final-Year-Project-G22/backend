package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
	libraryrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type libraryTemplateRepository struct {
	sharedrepo.GenericRepository[entity.LibraryTemplate]
	db     *core.Database
	logger core.Logger
}

func NewLibraryTemplateRepository(db *core.Database, logger core.Logger) libraryrepo.LibraryTemplateRepository {
	base := sharedrepo.NewBaseRepository[entity.LibraryTemplate](db, logger)
	return &libraryTemplateRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *libraryTemplateRepository) GetByGroupAndLanguage(ctx context.Context, groupID uuid.UUID, language string) (*entity.LibraryTemplate, error) {
	var tmpl entity.LibraryTemplate
	if err := getDB(ctx, r.db).
		Where("group_id = ? AND language = ?", groupID, language).
		First(&tmpl).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Error("Failed to get template by group and language", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &tmpl, nil
}

func (r *libraryTemplateRepository) ListByGroup(ctx context.Context, groupID uuid.UUID) ([]*entity.LibraryTemplate, error) {
	var templates []*entity.LibraryTemplate
	if err := getDB(ctx, r.db).
		Where("group_id = ?", groupID).
		Order("language asc").
		Find(&templates).Error; err != nil {
		r.logger.Error("Failed to list templates by group", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return templates, nil
}

func (r *libraryTemplateRepository) FindActiveByGroup(ctx context.Context, groupID uuid.UUID) ([]*entity.LibraryTemplate, error) {
	var templates []*entity.LibraryTemplate
	if err := getDB(ctx, r.db).
		Where("group_id = ? AND is_active = ?", groupID, true).
		Order("language asc").
		Find(&templates).Error; err != nil {
		r.logger.Error("Failed to find active templates by group", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return templates, nil
}
