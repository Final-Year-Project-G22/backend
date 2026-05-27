package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	return u.groupRepo.FindActive(ctx, q)
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

func (u *libraryViewUsecase) PreviewTemplate(ctx context.Context, input usecase.PreviewInput) (*usecase.PreviewOutput, error) {
	group, tmpl, err := u.resolveTemplate(ctx, input.GroupID, input.Language, input.AccountID)
	if err != nil {
		return nil, err
	}

	presignedURL, err := u.storage.GetPresignedURL(ctx, tmpl.FileKey, 5*time.Minute)
	if err != nil {
		u.logger.Error("Failed to generate preview URL", core.String("templateId", tmpl.ID.String()))
		return nil, apperrors.InternalError("library.errors.downloadFailed", err)
	}

	filename := fmt.Sprintf("%s_%s%s", group.Slug, tmpl.Language, u.extFromContentType(tmpl.ContentType))

	return &usecase.PreviewOutput{
		PresignedURL: presignedURL,
		ExpiresAt:    "5m",
		Filename:     filename,
		ContentType:  tmpl.ContentType,
	}, nil
}

func (u *libraryViewUsecase) DownloadTemplate(ctx context.Context, input usecase.DownloadInput) (*usecase.DownloadOutput, error) {
	group, tmpl, err := u.resolveTemplate(ctx, input.GroupID, input.Language, input.AccountID)
	if err != nil {
		return nil, err
	}

	presignedURL, err := u.storage.GetPresignedURL(ctx, tmpl.FileKey, 30*time.Minute)
	if err != nil {
		u.logger.Error("Failed to generate download URL", core.String("templateId", tmpl.ID.String()))
		return nil, apperrors.InternalError("library.errors.downloadFailed", err)
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
		PresignedURL: presignedURL,
		ExpiresAt:    "30m",
		Filename:     filename,
		ContentType:  tmpl.ContentType,
	}, nil
}

func (u *libraryViewUsecase) resolveTemplate(ctx context.Context, groupID uuid.UUID, language *string, accountID *uuid.UUID) (*entity.LibraryTemplateGroup, *entity.LibraryTemplate, error) {
	group, err := u.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, nil, err
	}

	if !group.IsActive {
		return nil, nil, apperrors.NotFoundErrorWithKey("library.errors.templateGroupNotFound")
	}

	if group.RequiresAuth && accountID == nil {
		return nil, nil, apperrors.UnauthorizedError("library.errors.authRequired")
	}

	if accountID != nil {
		allowed, err := u.tierSvc.HasAccess(ctx, *accountID, group.TierAccess)
		if err != nil {
			return nil, nil, err
		}
		if !allowed {
			return nil, nil, apperrors.ForbiddenError("library.errors.tierAccessDenied")
		}
	} else if group.TierAccess == entity.TierAccessPro {
		return nil, nil, apperrors.ForbiddenError("library.errors.tierAccessDenied")
	}

	lang := group.DefaultLanguage
	if language != nil && *language != "" {
		lang = *language
	}

	tmpl, err := u.tmplRepo.GetByGroupAndLanguage(ctx, group.ID, lang)
	if err != nil {
		return nil, nil, err
	}
	if tmpl == nil || !tmpl.IsActive {
		return nil, nil, apperrors.NotFoundErrorWithKey("library.errors.templateNotFound")
	}

	return group, tmpl, nil
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
