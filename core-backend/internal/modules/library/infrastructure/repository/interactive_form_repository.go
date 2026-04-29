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

type libraryInteractiveFormRepository struct {
	sharedrepo.GenericRepository[entity.LibraryInteractiveForm]
	db     *core.Database
	logger core.Logger
}

func NewLibraryInteractiveFormRepository(db *core.Database, logger core.Logger) libraryrepo.LibraryInteractiveFormRepository {
	base := sharedrepo.NewBaseRepository[entity.LibraryInteractiveForm](db, logger)
	return &libraryInteractiveFormRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *libraryInteractiveFormRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *libraryInteractiveFormRepository) GetByTemplateID(ctx context.Context, templateID uuid.UUID) (*entity.LibraryInteractiveForm, error) {
	var form entity.LibraryInteractiveForm
	if err := r.getDB(ctx).Where("template_id = ?", templateID).First(&form).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Error("Failed to get interactive form by template ID", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &form, nil
}
