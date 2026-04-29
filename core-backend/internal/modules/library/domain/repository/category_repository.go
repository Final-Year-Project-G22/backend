package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type LibraryCategoryRepository interface {
	sharedrepo.GenericRepository[entity.LibraryCategory]

	GetBySlug(ctx context.Context, parentID *uuid.UUID, slug string) (*entity.LibraryCategory, error)
	ListTree(ctx context.Context, includeInactive bool) ([]*entity.LibraryCategory, error)
	ListActive(ctx context.Context, q query.QueryOptions) ([]*entity.LibraryCategory, error)
	GetTranslations(ctx context.Context, categoryID uuid.UUID) ([]*entity.LibraryCategoryTranslation, error)
	UpsertTranslation(ctx context.Context, translation *entity.LibraryCategoryTranslation) error
	DeleteTranslation(ctx context.Context, categoryID uuid.UUID, language string) error
}
