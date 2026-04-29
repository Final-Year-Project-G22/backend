package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type LibraryInteractiveFormRepository interface {
	sharedrepo.GenericRepository[entity.LibraryInteractiveForm]

	GetByTemplateID(ctx context.Context, templateID uuid.UUID) (*entity.LibraryInteractiveForm, error)
}
