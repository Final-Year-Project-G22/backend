package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/google/uuid"
)

type LibraryTemplateRepository interface {
	sharedrepo.GenericRepository[entity.LibraryTemplate]

	GetByGroupAndLanguage(ctx context.Context, groupID uuid.UUID, language string) (*entity.LibraryTemplate, error)
	ListByGroup(ctx context.Context, groupID uuid.UUID) ([]*entity.LibraryTemplate, error)
	FindActiveByGroup(ctx context.Context, groupID uuid.UUID) ([]*entity.LibraryTemplate, error)
}
