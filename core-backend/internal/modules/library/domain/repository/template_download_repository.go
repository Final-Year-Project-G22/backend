package repository

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type LibraryTemplateDownloadRepository interface {
	sharedrepo.GenericRepository[entity.LibraryTemplateDownload]

	ListByAccount(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.LibraryTemplateDownload, error)
	CountByGroup(ctx context.Context, groupID uuid.UUID) (int64, error)
	ListAll(ctx context.Context, q query.QueryOptions) ([]*entity.LibraryTemplateDownload, error)
}
