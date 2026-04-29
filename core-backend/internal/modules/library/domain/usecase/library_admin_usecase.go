package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type LibraryAdminUsecase interface {
	CreateCategory(ctx context.Context, input CreateCategoryInput) (*entity.LibraryCategory, error)
	GetCategory(ctx context.Context, id uuid.UUID) (*entity.LibraryCategory, error)
	UpdateCategory(ctx context.Context, id uuid.UUID, input UpdateCategoryInput) (*entity.LibraryCategory, error)
	DeleteCategory(ctx context.Context, id uuid.UUID) error
	ListAllCategories(ctx context.Context, includeInactive bool) ([]*entity.LibraryCategory, error)
	AddCategoryTranslation(ctx context.Context, input CreateCategoryTranslationInput) (*entity.LibraryCategoryTranslation, error)
	UpdateCategoryTranslation(ctx context.Context, categoryID uuid.UUID, language string, input UpdateCategoryTranslationInput) (*entity.LibraryCategoryTranslation, error)
	DeleteCategoryTranslation(ctx context.Context, categoryID uuid.UUID, language string) error

	CreateTemplateGroup(ctx context.Context, createdBy uuid.UUID, input CreateTemplateGroupInput) (*entity.LibraryTemplateGroup, error)
	GetTemplateGroup(ctx context.Context, id uuid.UUID) (*entity.LibraryTemplateGroup, error)
	UpdateTemplateGroup(ctx context.Context, id uuid.UUID, input UpdateTemplateGroupInput) (*entity.LibraryTemplateGroup, error)
	DeleteTemplateGroup(ctx context.Context, id uuid.UUID) error
	ListAllTemplateGroups(ctx context.Context, categoryID *uuid.UUID, q query.QueryOptions) ([]*entity.LibraryTemplateGroup, error)

	CreateTemplate(ctx context.Context, input CreateTemplateInput) (*entity.LibraryTemplate, error)
	GetTemplate(ctx context.Context, id uuid.UUID) (*entity.LibraryTemplate, error)
	UpdateTemplate(ctx context.Context, id uuid.UUID, input UpdateTemplateInput) (*entity.LibraryTemplate, error)
	DeleteTemplate(ctx context.Context, id uuid.UUID) error

	CreateInteractiveForm(ctx context.Context, input CreateInteractiveFormInput) (*entity.LibraryInteractiveForm, error)
	GetInteractiveForm(ctx context.Context, id uuid.UUID) (*entity.LibraryInteractiveForm, error)
	UpdateInteractiveForm(ctx context.Context, id uuid.UUID, input UpdateInteractiveFormInput) (*entity.LibraryInteractiveForm, error)
	DeleteInteractiveForm(ctx context.Context, id uuid.UUID) error

	GetDownloadLogs(ctx context.Context, groupID *uuid.UUID, q query.QueryOptions) ([]*entity.LibraryTemplateDownload, error)
}
