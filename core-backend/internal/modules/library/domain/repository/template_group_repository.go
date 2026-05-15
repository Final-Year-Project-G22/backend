package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type LibraryTemplateGroupRepository interface {
	sharedrepo.GenericRepository[entity.LibraryTemplateGroup]

	GetBySlug(ctx context.Context, categoryID *uuid.UUID, slug string) (*entity.LibraryTemplateGroup, error)
	FindActive(ctx context.Context, q query.QueryOptions) ([]*entity.LibraryTemplateGroup, error)
	ListByCategory(ctx context.Context, categoryID uuid.UUID, q query.QueryOptions) ([]*entity.LibraryTemplateGroup, error)
	ListByFormat(ctx context.Context, format entity.TemplateFormat, q query.QueryOptions) ([]*entity.LibraryTemplateGroup, error)
	IncrementDownloadCount(ctx context.Context, id uuid.UUID) error
}
