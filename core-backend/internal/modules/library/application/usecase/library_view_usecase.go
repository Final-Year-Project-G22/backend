package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
	libraryerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/error"
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

func (u *libraryViewUsecase) GetTemplateGroup(ctx context.Context, slug string, locale *string) (*entity.LibraryTemplateGroup, []*entity.LibraryTemplate, error) {
	group, err := u.groupRepo.GetBySlug(ctx, nil, slug)
	if err != nil {
		return nil, nil, err
	}

	if !group.IsActive {
		return nil, nil, libraryerror.ErrTemplateGroupNotFound
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
	group, err := u.groupRepo.GetBySlug(ctx, nil, input.Slug)
	if err != nil {
		return nil, err
	}

	if !group.IsActive {
		return nil, libraryerror.ErrTemplateGroupNotFound
	}

	if group.RequiresAuth && input.AccountID == nil {
		return nil, libraryerror.ErrAuthRequired
	}

	if input.AccountID != nil {
		allowed, err := u.tierSvc.HasAccess(ctx, *input.AccountID, group.TierAccess)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, libraryerror.ErrTierAccessDenied
		}
	} else if group.TierAccess == entity.TierAccessPro {
		return nil, libraryerror.ErrTierAccessDenied
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
		return nil, libraryerror.ErrTemplateNotFound
	}

	presignedURL, err := u.storage.GetPresignedURL(ctx, tmpl.FileKey, 5*time.Minute)
	if err != nil {
		u.logger.Error("Failed to generate presigned URL", core.Error(err))
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
