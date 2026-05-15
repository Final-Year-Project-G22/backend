package usecase

import (
	"context"
	"errors"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
	libraryerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/usecase"
	sharedRepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type libraryAdminUsecase struct {
	catRepo      repository.LibraryCategoryRepository
	groupRepo    repository.LibraryTemplateGroupRepository
	tmplRepo     repository.LibraryTemplateRepository
	formRepo     repository.LibraryInteractiveFormRepository
	downloadRepo repository.LibraryTemplateDownloadRepository
	transactor   sharedRepo.Transactor
	logger       core.Logger
}

func NewLibraryAdminUsecase(
	catRepo repository.LibraryCategoryRepository,
	groupRepo repository.LibraryTemplateGroupRepository,
	tmplRepo repository.LibraryTemplateRepository,
	formRepo repository.LibraryInteractiveFormRepository,
	downloadRepo repository.LibraryTemplateDownloadRepository,
	transactor sharedRepo.Transactor,
	logger core.Logger,
) usecase.LibraryAdminUsecase {
	return &libraryAdminUsecase{
		catRepo:      catRepo,
		groupRepo:    groupRepo,
		tmplRepo:     tmplRepo,
		formRepo:     formRepo,
		downloadRepo: downloadRepo,
		transactor:   transactor,
		logger:       logger,
	}
}

func (u *libraryAdminUsecase) CreateCategory(ctx context.Context, input usecase.CreateCategoryInput) (*entity.LibraryCategory, error) {
	existing, err := u.catRepo.GetBySlug(ctx, input.ParentCategoryID, input.Slug)
	if err != nil && !errors.Is(err, libraryerror.ErrCategoryNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, apperrors.AlreadyExistsError("category", "slug", input.Slug)
	}

	cat := &entity.LibraryCategory{
		Name:             input.Name,
		Slug:             input.Slug,
		Icon:             input.Icon,
		SortOrder:        input.SortOrder,
		ParentCategoryID: input.ParentCategoryID,
		IsActive:         true,
	}
	if err := u.catRepo.Create(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (u *libraryAdminUsecase) GetCategory(ctx context.Context, id uuid.UUID) (*entity.LibraryCategory, error) {
	cat, err := u.catRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	translations, err := u.catRepo.GetTranslations(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, t := range translations {
		cat.Translations = append(cat.Translations, *t)
	}
	return cat, nil
}

func (u *libraryAdminUsecase) UpdateCategory(ctx context.Context, id uuid.UUID, input usecase.UpdateCategoryInput) (*entity.LibraryCategory, error) {
	cat, err := u.catRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		cat.Name = *input.Name
	}
	if input.Slug != nil {
		existing, err := u.catRepo.GetBySlug(ctx, cat.ParentCategoryID, *input.Slug)
		if err != nil && !errors.Is(err, libraryerror.ErrCategoryNotFound) {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, apperrors.AlreadyExistsError("category", "slug", *input.Slug)
		}
		cat.Slug = *input.Slug
	}
	if input.Icon != nil {
		cat.Icon = input.Icon
	}
	if input.SortOrder != nil {
		cat.SortOrder = *input.SortOrder
	}
	if input.IsActive != nil {
		cat.IsActive = *input.IsActive
	}
	if err := u.catRepo.Update(ctx, cat); err != nil {
		return nil, err
	}
	translations, _ := u.catRepo.GetTranslations(ctx, id)
	for _, t := range translations {
		cat.Translations = append(cat.Translations, *t)
	}
	return cat, nil
}

func (u *libraryAdminUsecase) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	groups, err := u.groupRepo.ListByCategory(ctx, id, query.QueryOptions{Page: 1, PageSize: 1})
	if err != nil {
		return err
	}
	if len(groups) > 0 {
		return libraryerror.ErrCategoryHasActiveGroups
	}
	return u.catRepo.Delete(ctx, id)
}

func (u *libraryAdminUsecase) ListAllCategories(ctx context.Context, includeInactive bool) ([]*entity.LibraryCategory, error) {
	return u.catRepo.ListTree(ctx, includeInactive)
}

func (u *libraryAdminUsecase) AddCategoryTranslation(ctx context.Context, input usecase.CreateCategoryTranslationInput) (*entity.LibraryCategoryTranslation, error) {
	if _, err := u.catRepo.GetByID(ctx, input.CategoryID); err != nil {
		return nil, err
	}
	trans := &entity.LibraryCategoryTranslation{
		LibraryCategoryID: input.CategoryID,
		Language:          input.Language,
		Name:              input.Name,
		Description:       input.Description,
	}
	if err := u.catRepo.UpsertTranslation(ctx, trans); err != nil {
		return nil, err
	}
	return trans, nil
}

func (u *libraryAdminUsecase) UpdateCategoryTranslation(ctx context.Context, categoryID uuid.UUID, language string, input usecase.UpdateCategoryTranslationInput) (*entity.LibraryCategoryTranslation, error) {
	translations, err := u.catRepo.GetTranslations(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	var target *entity.LibraryCategoryTranslation
	for _, t := range translations {
		if t.Language == language {
			target = t
			break
		}
	}
	if target == nil {
		return nil, apperrors.NotFoundError("categoryTranslation", language)
	}
	if input.Name != nil {
		target.Name = *input.Name
	}
	if input.Description != nil {
		target.Description = input.Description
	}
	if err := u.catRepo.UpsertTranslation(ctx, target); err != nil {
		return nil, err
	}
	return target, nil
}

func (u *libraryAdminUsecase) DeleteCategoryTranslation(ctx context.Context, categoryID uuid.UUID, language string) error {
	return u.catRepo.DeleteTranslation(ctx, categoryID, language)
}

func (u *libraryAdminUsecase) CreateTemplateGroup(ctx context.Context, createdBy uuid.UUID, input usecase.CreateTemplateGroupInput) (*entity.LibraryTemplateGroup, error) {
	existing, err := u.groupRepo.GetBySlug(ctx, &input.CategoryID, input.Slug)
	if err != nil && !errors.Is(err, libraryerror.ErrTemplateGroupNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, apperrors.AlreadyExistsError("templateGroup", "slug", input.Slug)
	}

	group := &entity.LibraryTemplateGroup{
		Name:            input.Name,
		Description:     input.Description,
		Slug:            input.Slug,
		CategoryID:      input.CategoryID,
		Format:          input.Format,
		TierAccess:      input.TierAccess,
		RequiresAuth:    input.RequiresAuth,
		SortOrder:       input.SortOrder,
		DefaultLanguage: input.DefaultLanguage,
		ThumbnailURL:    input.ThumbnailURL,
		IsActive:        true,
		CreatedBy:       createdBy,
	}
	if err := u.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}

	if err := u.groupRepo.UpdateByID(ctx, group.ID, map[string]interface{}{"is_active": false}); err != nil {
		return nil, err
	}
	group.IsActive = false

	return group, nil
}

func (u *libraryAdminUsecase) GetTemplateGroup(ctx context.Context, id uuid.UUID) (*entity.LibraryTemplateGroup, error) {
	return u.groupRepo.GetByID(ctx, id)
}

func (u *libraryAdminUsecase) UpdateTemplateGroup(ctx context.Context, id uuid.UUID, input usecase.UpdateTemplateGroupInput) (*entity.LibraryTemplateGroup, error) {
	group, err := u.groupRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		group.Name = *input.Name
	}
	if input.Description != nil {
		group.Description = input.Description
	}
	if input.Slug != nil {
		existing, err := u.groupRepo.GetBySlug(ctx, &group.CategoryID, *input.Slug)
		if err != nil && !errors.Is(err, libraryerror.ErrTemplateGroupNotFound) {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, apperrors.AlreadyExistsError("templateGroup", "slug", *input.Slug)
		}
		group.Slug = *input.Slug
	}
	if input.CategoryID != nil {
		group.CategoryID = *input.CategoryID
	}
	if input.Format != nil {
		group.Format = *input.Format
	}
	if input.TierAccess != nil {
		group.TierAccess = *input.TierAccess
	}
	if input.RequiresAuth != nil {
		group.RequiresAuth = *input.RequiresAuth
	}
	if input.SortOrder != nil {
		group.SortOrder = *input.SortOrder
	}
	if input.DefaultLanguage != nil {
		group.DefaultLanguage = *input.DefaultLanguage
	}
	if input.IsActive != nil {
		group.IsActive = *input.IsActive
	}
	if input.ThumbnailURL != nil {
		group.ThumbnailURL = input.ThumbnailURL
	}
	if err := u.groupRepo.Update(ctx, group); err != nil {
		return nil, err
	}
	return group, nil
}

func (u *libraryAdminUsecase) DeleteTemplateGroup(ctx context.Context, id uuid.UUID) error {
	return u.groupRepo.Delete(ctx, id)
}

func (u *libraryAdminUsecase) ListAllTemplateGroups(ctx context.Context, categoryID *uuid.UUID, q query.QueryOptions) ([]*entity.LibraryTemplateGroup, int64, error) {
	if categoryID != nil {
		groups, err := u.groupRepo.ListByCategory(ctx, *categoryID, q)
		if err != nil {
			return nil, 0, err
		}
		var total int64
		if err := u.groupRepo.GetDB().Model(&entity.LibraryTemplateGroup{}).
			Where("category_id = ?", *categoryID).
			Count(&total).Error; err != nil {
			return nil, 0, err
		}
		return groups, total, nil
	}
	result := u.groupRepo.FindAll(ctx, q)
	return result.Data, result.Total, nil
}

func (u *libraryAdminUsecase) CreateTemplate(ctx context.Context, input usecase.CreateTemplateInput) (*entity.LibraryTemplate, error) {
	group, err := u.groupRepo.GetByID(ctx, input.GroupID)
	if err != nil {
		return nil, err
	}

	existing, err := u.tmplRepo.GetByGroupAndLanguage(ctx, input.GroupID, input.Language)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperrors.AlreadyExistsError("template", "language", input.Language)
	}

	tmpl := &entity.LibraryTemplate{
		GroupID:     input.GroupID,
		Language:    input.Language,
		Title:       input.Title,
		Description: input.Description,
		FileKey:     input.FileKey,
		FileURL:     input.FileURL,
		FileSize:    input.FileSize,
		ContentType: input.ContentType,
		Version:     1,
		IsActive:    true,
	}
	if err := u.tmplRepo.Create(ctx, tmpl); err != nil {
		return nil, err
	}

	if !group.IsActive {
		if err := u.groupRepo.UpdateByID(ctx, group.ID, map[string]interface{}{"is_active": true}); err != nil {
			return nil, err
		}
	}

	return tmpl, nil
}

func (u *libraryAdminUsecase) GetTemplate(ctx context.Context, id uuid.UUID) (*entity.LibraryTemplate, error) {
	return u.tmplRepo.GetByID(ctx, id)
}

func (u *libraryAdminUsecase) UpdateTemplate(ctx context.Context, id uuid.UUID, input usecase.UpdateTemplateInput) (*entity.LibraryTemplate, error) {
	tmpl, err := u.tmplRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Title != nil {
		tmpl.Title = *input.Title
	}
	if input.Description != nil {
		tmpl.Description = input.Description
	}
	if input.IsActive != nil {
		tmpl.IsActive = *input.IsActive
	}
	if input.FileBytes != nil {
		tmpl.Version++
	}
	if input.FileKey != nil {
		tmpl.FileKey = *input.FileKey
	}
	if input.FileURL != nil {
		tmpl.FileURL = input.FileURL
	}
	if input.FileSize != nil {
		tmpl.FileSize = *input.FileSize
	}
	if input.ContentType != nil {
		tmpl.ContentType = *input.ContentType
	}
	if err := u.tmplRepo.Update(ctx, tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (u *libraryAdminUsecase) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	return u.tmplRepo.Delete(ctx, id)
}

func (u *libraryAdminUsecase) CreateInteractiveForm(ctx context.Context, input usecase.CreateInteractiveFormInput) (*entity.LibraryInteractiveForm, error) {
	tmpl, err := u.tmplRepo.GetByID(ctx, input.TemplateID)
	if err != nil {
		return nil, err
	}

	group, err := u.groupRepo.GetByID(ctx, tmpl.GroupID)
	if err != nil {
		return nil, err
	}
	if group.Format != entity.TemplateFormatInteractiveForm {
		return nil, apperrors.InvalidInputError("templateId", "library.errors.invalidFileType")
	}

	if tmpl.InteractiveForm != nil {
		return nil, apperrors.AlreadyExistsError("interactiveForm", "templateID", input.TemplateID)
	}

	form := &entity.LibraryInteractiveForm{
		TemplateID:  input.TemplateID,
		Name:        input.Name,
		Description: input.Description,
		FormLayout:  input.FormLayout,
		Version:     1,
		IsActive:    true,
	}
	if err := u.formRepo.Create(ctx, form); err != nil {
		return nil, err
	}
	return form, nil
}

func (u *libraryAdminUsecase) GetInteractiveForm(ctx context.Context, id uuid.UUID) (*entity.LibraryInteractiveForm, error) {
	return u.formRepo.GetByID(ctx, id)
}

func (u *libraryAdminUsecase) GetInteractiveFormByTemplate(ctx context.Context, templateID uuid.UUID) (*entity.LibraryInteractiveForm, error) {
	form, err := u.formRepo.GetByTemplateID(ctx, templateID)
	if err != nil {
		return nil, err
	}
	if form == nil {
		return nil, apperrors.NotFoundError("interactiveForm", templateID)
	}
	return form, nil
}

func (u *libraryAdminUsecase) UpdateInteractiveForm(ctx context.Context, id uuid.UUID, input usecase.UpdateInteractiveFormInput) (*entity.LibraryInteractiveForm, error) {
	form, err := u.formRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		form.Name = *input.Name
	}
	if input.Description != nil {
		form.Description = input.Description
	}
	if input.FormLayout != nil {
		form.FormLayout = *input.FormLayout
		form.Version++
	}
	if err := u.formRepo.Update(ctx, form); err != nil {
		return nil, err
	}
	return form, nil
}

func (u *libraryAdminUsecase) DeleteInteractiveForm(ctx context.Context, id uuid.UUID) error {
	return u.formRepo.Delete(ctx, id)
}

func (u *libraryAdminUsecase) GetDownloadLogs(ctx context.Context, groupID *uuid.UUID, q query.QueryOptions) ([]*entity.LibraryTemplateDownload, error) {
	if groupID != nil {
		var downloads []*entity.LibraryTemplateDownload
		all, err := u.downloadRepo.ListAll(ctx, q)
		if err != nil {
			return nil, err
		}
		for _, d := range all {
			if d.GroupID == *groupID {
				downloads = append(downloads, d)
			}
		}
		return downloads, nil
	}
	return u.downloadRepo.ListAll(ctx, q)
}
