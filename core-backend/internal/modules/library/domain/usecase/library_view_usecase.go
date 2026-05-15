package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type LibraryViewUsecase interface {
	ListCategories(ctx context.Context, locale *string) ([]*entity.LibraryCategory, error)
	ListTemplateGroups(ctx context.Context, categoryID *uuid.UUID, format *entity.TemplateFormat, q query.QueryOptions) ([]*entity.LibraryTemplateGroup, error)
	GetTemplateGroup(ctx context.Context, groupID uuid.UUID, locale *string) (*entity.LibraryTemplateGroup, []*entity.LibraryTemplate, error)
	PreviewTemplate(ctx context.Context, input PreviewInput) (*PreviewOutput, error)
	DownloadTemplate(ctx context.Context, input DownloadInput) (*DownloadOutput, error)
	ListMyDownloads(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.LibraryTemplateDownload, error)
}
