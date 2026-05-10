package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/usecase"
	sharedRepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/Final-Year-Project-G22/backend/core/pkg/storage"
	"github.com/google/uuid"
)

type libraryViewUsecase struct {
	catRepo      repository.LibraryCategoryRepository
	groupRepo    repository.LibraryTemplateGroupRepository
	tmplRepo     repository.LibraryTemplateRepository
	downloadRepo repository.LibraryTemplateDownloadRepository
	tierSvc      usecase.TierService
	storage      storage.Storage
	transactor   sharedRepo.Transactor
	logger       core.Logger
}

func NewLibraryViewUsecase(
	catRepo repository.LibraryCategoryRepository,
	groupRepo repository.LibraryTemplateGroupRepository,
	tmplRepo repository.LibraryTemplateRepository,
	downloadRepo repository.LibraryTemplateDownloadRepository,
	tierSvc usecase.TierService,
	storage storage.Storage,
	transactor sharedRepo.Transactor,
	logger core.Logger,
) usecase.LibraryViewUsecase {
	return &libraryViewUsecase{
		catRepo:      catRepo,
		groupRepo:    groupRepo,
		tmplRepo:     tmplRepo,
		downloadRepo: downloadRepo,
		tierSvc:      tierSvc,
		storage:      storage,
		transactor:   transactor,
		logger:       logger,
	}
}

func (u *libraryViewUsecase) ListCategories(ctx context.Context, locale *string) ([]*entity.LibraryCategory, error) {
	categories, err := u.catRepo.ListTree(ctx, false)
	if err != nil {
		return nil, err
	}

	if locale != nil && *locale != "" {
		for _, cat := range categories {
			translations, err := u.catRepo.GetTranslations(ctx, cat.ID)
			if err != nil {
				return nil, err
			}
			for _, t := range translations {
				if t.Language == *locale {
					cat.Name = t.Name
					break
				}
			}
		}
	}

	return categories, nil
}

func (u *libraryViewUsecase) ListTemplateGroups(ctx context.Context, categoryID *uuid.UUID, format *entity.TemplateFormat, q query.QueryOptions) ([]*entity.LibraryTemplateGroup, error) {
	if categoryID != nil {
		return u.groupRepo.ListByCategory(ctx, *categoryID, q)
	}
	if format != nil {
		return u.groupRepo.ListByFormat(ctx, *format, q)
	}
	return u.groupRepo.Find(ctx, q)
}

func (u *libraryViewUsecase) GetTemplateGroup(ctx context.Context, groupID uuid.UUID, locale *string) (*entity.LibraryTemplateGroup, []*entity.LibraryTemplate, error) {
	group, err := u.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, nil, err
	}

	if !group.IsActive {
		return nil, nil, apperrors.NotFoundErrorWithKey("library.errors.templateGroupNotFound")
	}

	templates, err := u.tmplRepo.FindActiveByGroup(ctx, group.ID)
	if err != nil {
		return nil, nil, err
	}

	if locale != nil && *locale != "" {
		found := false
		for _, t := range templates {
			if t.Language == *locale {
				found = true
				break
			}
		}
		if !found && group.DefaultLanguage != "" {
			for _, t := range templates {
				if t.Language == group.DefaultLanguage {
					matched := []*entity.LibraryTemplate{t}
					return group, matched, nil
				}
			}
		}
	}

	return group, templates, nil
}

func (u *libraryViewUsecase) DownloadTemplate(ctx context.Context, input usecase.DownloadInput) (*usecase.DownloadOutput, error) {
	group, err := u.groupRepo.GetByID(ctx, input.GroupID)
	if err != nil {
		return nil, err
	}

	if !group.IsActive {
		return nil, apperrors.NotFoundErrorWithKey("library.errors.templateGroupNotFound")
	}

	if group.RequiresAuth && input.AccountID == nil {
		return nil, apperrors.UnauthorizedError("library.errors.authRequired")
	}

	if input.AccountID != nil {
		allowed, err := u.tierSvc.HasAccess(ctx, *input.AccountID, group.TierAccess)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, apperrors.ForbiddenError("library.errors.tierAccessDenied")
		}
	} else if group.TierAccess == entity.TierAccessPro {
		return nil, apperrors.ForbiddenError("library.errors.tierAccessDenied")
	}

	lang := group.DefaultLanguage
	if input.Language != nil && *input.Language != "" {
		lang = *input.Language
	}

	tmpl, err := u.tmplRepo.GetByGroupAndLanguage(ctx, group.ID, lang)
	if err != nil {
		return nil, err
	}
	if tmpl == nil || !tmpl.IsActive {
		return nil, apperrors.NotFoundErrorWithKey("library.errors.templateNotFound")
	}

	downloadURL := ""
	if tmpl.FileURL != nil {
		downloadURL = *tmpl.FileURL
	}
	if downloadURL == "" {
		u.logger.Error("No download URL available for template", core.String("templateId", tmpl.ID.String()))
		return nil, apperrors.InternalError("library.errors.downloadFailed", nil)
	}

	filename := fmt.Sprintf("%s_%s%s", group.Slug, tmpl.Language, u.extFromContentType(tmpl.ContentType))

	err = u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := u.groupRepo.IncrementDownloadCount(txCtx, group.ID); err != nil {
			return err
		}

		if input.AccountID != nil {
			dl := &entity.LibraryTemplateDownload{
				AccountID:  *input.AccountID,
				TemplateID: tmpl.ID,
				GroupID:    group.ID,
			}
			if err := u.downloadRepo.Create(txCtx, dl); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &usecase.DownloadOutput{
		PresignedURL: downloadURL,
		ExpiresAt:    "5m",
		Filename:     filename,
	}, nil
}

func (u *libraryViewUsecase) extFromContentType(contentType string) string {
	switch {
	case strings.Contains(contentType, "pdf"):
		return ".pdf"
	case strings.Contains(contentType, "wordprocessingml"):
		return ".docx"
	case strings.Contains(contentType, "spreadsheetml"):
		return ".xlsx"
	default:
		ext := strings.Split(contentType, "/")
		if len(ext) == 2 {
			return "." + ext[1]
		}
		return ""
	}
}

func (u *libraryViewUsecase) ListMyDownloads(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.LibraryTemplateDownload, error) {
	return u.downloadRepo.ListByAccount(ctx, accountID, q)
}
