package handler

import (
	"context"
	"io"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/i18n"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

const maxUploadSize int64 = 10 * 1024 * 1024

type LibraryAdminHandler struct {
	adminUC usecase.LibraryAdminUsecase
	svc     service.LibraryService
}

func NewLibraryAdminHandler(adminUC usecase.LibraryAdminUsecase, svc service.LibraryService) *LibraryAdminHandler {
	return &LibraryAdminHandler{adminUC: adminUC, svc: svc}
}

// --- Categories ---

func (h *LibraryAdminHandler) HandleCreateCategory(ctx context.Context, input *dto.CreateLibraryCategoryInput) (*dto.CreateLibraryCategoryOutput, error) {
	cat, err := h.adminUC.CreateCategory(ctx, usecase.CreateCategoryInput{
		Name:             input.Body.Name,
		Slug:             input.Body.Slug,
		Icon:             input.Body.Icon,
		SortOrder:        input.Body.SortOrder,
		ParentCategoryID: input.Body.ParentCategoryID,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.CreateLibraryCategoryOutput{Body: struct {
		ID uuid.UUID `json:"id" doc:"Created category ID"`
	}{ID: cat.ID}}, nil
}

func (h *LibraryAdminHandler) HandleGetCategory(ctx context.Context, input *dto.LibraryGetCategoryInput) (*dto.LibraryGetCategoryOutput, error) {
	cat, err := h.adminUC.GetCategory(ctx, input.ID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.LibraryGetCategoryOutput{Body: dto.ToCategoryDetailResponse(cat)}, nil
}

func (h *LibraryAdminHandler) HandleUpdateCategory(ctx context.Context, input *dto.UpdateLibraryCategoryInput) (*dto.UpdateLibraryCategoryOutput, error) {
	ucInput := usecase.UpdateCategoryInput{
		Name:      input.Body.Name,
		Slug:      input.Body.Slug,
		Icon:      input.Body.Icon,
		SortOrder: input.Body.SortOrder,
		IsActive:  input.Body.IsActive,
	}
	cat, err := h.adminUC.UpdateCategory(ctx, input.ID, ucInput)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UpdateLibraryCategoryOutput{Body: dto.ToCategoryDetailResponse(cat)}, nil
}

func (h *LibraryAdminHandler) HandleDeleteCategory(ctx context.Context, input *dto.LibraryDeleteCategoryInput) (*dto.LibraryDeleteCategoryOutput, error) {
	if err := h.adminUC.DeleteCategory(ctx, input.ID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.LibraryDeleteCategoryOutput{Body: struct {
		Message string `json:"message" doc:"Success message"`
	}{Message: i18n.T(ctx, "library.successes.categoryDeleted")}}, nil
}

func (h *LibraryAdminHandler) HandleListAllCategories(ctx context.Context, input *dto.ListAllCategoriesInput) (*dto.ListAllCategoriesOutput, error) {
	categories, err := h.adminUC.ListAllCategories(ctx, input.IncludeInactive)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	resp := make([]dto.CategorySummaryResponse, len(categories))
	for i, c := range categories {
		resp[i] = dto.ToCategorySummaryResponse(c)
	}
	return &dto.ListAllCategoriesOutput{Body: resp}, nil
}

// --- Category Translations ---

func (h *LibraryAdminHandler) HandleAddCategoryTranslation(ctx context.Context, input *dto.AddCategoryTranslationInput) (*dto.AddCategoryTranslationOutput, error) {
	trans, err := h.adminUC.AddCategoryTranslation(ctx, usecase.CreateCategoryTranslationInput{
		CategoryID:  input.ID,
		Language:    input.Body.Language,
		Name:        input.Body.Name,
		Description: input.Body.Description,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.AddCategoryTranslationOutput{Body: struct {
		ID uuid.UUID `json:"id" doc:"Translation ID"`
	}{ID: trans.ID}}, nil
}

func (h *LibraryAdminHandler) HandleUpdateCategoryTranslation(ctx context.Context, input *dto.UpdateCategoryTranslationInput) (*dto.UpdateCategoryTranslationOutput, error) {
	_, err := h.adminUC.UpdateCategoryTranslation(ctx, input.CategoryID, input.Language, usecase.UpdateCategoryTranslationInput{
		Name:        input.Body.Name,
		Description: input.Body.Description,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UpdateCategoryTranslationOutput{Body: struct {
		Message string `json:"message" doc:"Success message"`
	}{Message: i18n.T(ctx, "library.successes.translationUpdated")}}, nil
}

func (h *LibraryAdminHandler) HandleDeleteCategoryTranslation(ctx context.Context, input *dto.DeleteCategoryTranslationInput) (*dto.DeleteCategoryTranslationOutput, error) {
	if err := h.adminUC.DeleteCategoryTranslation(ctx, input.CategoryID, input.Language); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.DeleteCategoryTranslationOutput{Body: struct {
		Message string `json:"message" doc:"Success message"`
	}{Message: i18n.T(ctx, "library.successes.translationDeleted")}}, nil
}

// --- Template Groups ---

func (h *LibraryAdminHandler) HandleCreateTemplateGroup(ctx context.Context, input *dto.CreateTemplateGroupInput) (*dto.CreateTemplateGroupOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if accountID == uuid.Nil {
		return nil, apperrors.ToHumaError(ctx, apperrors.UnauthorizedError("library.errors.authRequired"))
	}

	group, err := h.svc.CreateTemplateGroup(ctx, accountID, usecase.CreateTemplateGroupInput{
		Name:            input.Body.Name,
		Description:     input.Body.Description,
		Slug:            input.Body.Slug,
		CategoryID:      input.Body.CategoryID,
		Format:          input.Body.Format,
		TierAccess:      input.Body.TierAccess,
		RequiresAuth:    input.Body.RequiresAuth,
		SortOrder:       input.Body.SortOrder,
		DefaultLanguage: input.Body.DefaultLanguage,
		ThumbnailURL:    input.Body.ThumbnailURL,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.CreateTemplateGroupOutput{Body: struct {
		ID uuid.UUID `json:"id" doc:"Created group ID"`
	}{ID: group.ID}}, nil
}

func (h *LibraryAdminHandler) HandleGetTemplateGroup(ctx context.Context, input *dto.GetTemplateGroupInput) (*dto.GetTemplateGroupOutput, error) {
	group, err := h.adminUC.GetTemplateGroup(ctx, input.GroupID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.GetTemplateGroupOutput{Body: dto.ToTemplateGroupDetailResponse(group)}, nil
}

func (h *LibraryAdminHandler) HandleUpdateTemplateGroup(ctx context.Context, input *dto.UpdateTemplateGroupInput) (*dto.UpdateTemplateGroupOutput, error) {
	ucInput := usecase.UpdateTemplateGroupInput{
		Name:            input.Body.Name,
		Description:     input.Body.Description,
		Slug:            input.Body.Slug,
		CategoryID:      input.Body.CategoryID,
		Format:          input.Body.Format,
		TierAccess:      input.Body.TierAccess,
		RequiresAuth:    input.Body.RequiresAuth,
		SortOrder:       input.Body.SortOrder,
		DefaultLanguage: input.Body.DefaultLanguage,
		IsActive:        input.Body.IsActive,
		ThumbnailURL:    input.Body.ThumbnailURL,
	}
	group, err := h.svc.UpdateTemplateGroup(ctx, input.GroupID, ucInput)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UpdateTemplateGroupOutput{Body: dto.ToTemplateGroupDetailResponse(group)}, nil
}

func (h *LibraryAdminHandler) HandleDeleteTemplateGroup(ctx context.Context, input *dto.DeleteTemplateGroupInput) (*dto.DeleteTemplateGroupOutput, error) {
	if err := h.adminUC.DeleteTemplateGroup(ctx, input.GroupID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.DeleteTemplateGroupOutput{Body: struct {
		Message string `json:"message" doc:"Success message"`
	}{Message: i18n.T(ctx, "library.successes.templateGroupDeleted")}}, nil
}

func (h *LibraryAdminHandler) HandleListAllTemplateGroups(ctx context.Context, input *dto.ListAllTemplateGroupsInput) (*dto.ListAllTemplateGroupsOutput, error) {
	q := dto.ToQueryOptions(input.Page, input.PageSize)
	var categoryID *uuid.UUID
	if input.CategoryID != "" {
		id, err := uuid.Parse(input.CategoryID)
		if err != nil {
			return nil, apperrors.ToHumaError(ctx, apperrors.InvalidInputError("categoryId", "library.errors.invalidFileType"))
		}
		categoryID = &id
	}
	groups, total, err := h.adminUC.ListAllTemplateGroups(ctx, categoryID, q)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	data := make([]dto.TemplateGroupSummaryResponse, len(groups))
	for i, g := range groups {
		data[i] = dto.ToTemplateGroupSummaryResponse(g)
	}
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	return &dto.ListAllTemplateGroupsOutput{Body: dto.ListAllTemplateGroupsResponseBody{
		Data:       data,
		Total:      total,
		Page:       q.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}}, nil
}

// --- Templates ---

func (h *LibraryAdminHandler) HandleCreateTemplateUploadIntent(ctx context.Context, input *dto.LibraryCreateUploadIntentInput) (*dto.LibraryCreateUploadIntentOutput, error) {
	out, err := h.svc.CreateUploadIntent(ctx, usecase.CreateTemplateUploadIntentInput{
		GroupID:     input.GroupID,
		Language:    input.Body.Language,
		Title:       input.Body.Title,
		Description: input.Body.Description,
		FileName:    input.Body.FileName,
		ContentType: input.Body.ContentType,
		FileSize:    input.Body.FileSize,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.LibraryCreateUploadIntentOutput{Body: struct {
		UploadURL string            `json:"uploadUrl" doc:"Direct upload URL"`
		Method    string            `json:"method" doc:"HTTP method for upload"`
		Headers   map[string]string `json:"headers,omitempty" doc:"Required upload headers"`
		FileKey   string            `json:"fileKey" doc:"Storage key for the file"`
		ExpiresAt string            `json:"expiresAt" doc:"Upload URL expiry time"`
	}{
		UploadURL: out.UploadURL,
		Method:    out.Method,
		Headers:   out.Headers,
		FileKey:   out.FileKey,
		ExpiresAt: out.ExpiresAt,
	}}, nil
}

func (h *LibraryAdminHandler) HandleCreateTemplate(ctx context.Context, input *dto.LibraryCreateTemplateInput) (*dto.LibraryCreateTemplateOutput, error) {
	var desc *string
	if input.Body.Description != nil && *input.Body.Description != "" {
		desc = input.Body.Description
	}
	tmpl, err := h.svc.CreateTemplate(ctx, usecase.CreateTemplateInput{
		GroupID:     input.GroupID,
		Language:    input.Body.Language,
		Title:       input.Body.Title,
		Description: desc,
		FileKey:     input.Body.FileKey,
		FileSize:    input.Body.FileSize,
		ContentType: input.Body.ContentType,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.LibraryCreateTemplateOutput{Body: struct {
		ID uuid.UUID `json:"id" doc:"Created template ID"`
	}{ID: tmpl.ID}}, nil
}

func (h *LibraryAdminHandler) HandleGetTemplate(ctx context.Context, input *dto.LibraryGetTemplateInput) (*dto.LibraryGetTemplateOutput, error) {
	tmpl, err := h.adminUC.GetTemplate(ctx, input.TemplateID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.LibraryGetTemplateOutput{Body: dto.ToLibraryTemplateDetailResponse(tmpl)}, nil
}

func (h *LibraryAdminHandler) HandleUpdateTemplate(ctx context.Context, input *dto.LibraryUpdateTemplateInput) (*dto.LibraryUpdateTemplateOutput, error) {
	data := input.RawBody.Data()
	if data == nil {
		return nil, apperrors.ToHumaError(ctx, apperrors.BadRequestError("library.errors.invalidFile"))
	}

	var title *string
	if data.Title != "" {
		title = &data.Title
	}
	var desc *string
	if data.Description != "" {
		desc = &data.Description
	}
	var isActive *bool
	switch data.IsActive {
	case "true":
		t := true
		isActive = &t
	case "false":
		f := false
		isActive = &f
	}
	ucInput := usecase.UpdateTemplateInput{
		Title:       title,
		Description: desc,
		IsActive:    isActive,
	}

	if data.File.IsSet {
		file := data.File.File
		if file != nil {
			defer func() { _ = file.Close() }()
			limitedReader := io.LimitReader(file, maxUploadSize+1)
			fileBytes, err := io.ReadAll(limitedReader)
			if err != nil {
				return nil, apperrors.ToHumaError(ctx, apperrors.InternalError("library.errors.readFileFailed", err))
			}
			if int64(len(fileBytes)) > maxUploadSize {
				return nil, apperrors.ToHumaError(ctx, apperrors.PayloadTooLargeError("library.errors.fileTooLarge"))
			}
			ucInput.FileBytes = fileBytes
			ucInput.Filename = &input.RawBody.Data().File.Filename
		}
	}

	tmpl, err := h.svc.UpdateTemplate(ctx, input.TemplateID, ucInput)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.LibraryUpdateTemplateOutput{Body: dto.ToLibraryTemplateDetailResponse(tmpl)}, nil
}

func (h *LibraryAdminHandler) HandleDeleteTemplate(ctx context.Context, input *dto.LibraryDeleteTemplateInput) (*dto.LibraryDeleteTemplateOutput, error) {
	if err := h.adminUC.DeleteTemplate(ctx, input.TemplateID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.LibraryDeleteTemplateOutput{Body: struct {
		Message string `json:"message" doc:"Success message"`
	}{Message: i18n.T(ctx, "library.successes.templateDeleted")}}, nil
}

func (h *LibraryAdminHandler) HandleListTemplatesByGroup(ctx context.Context, input *dto.ListTemplatesByGroupInput) (*dto.ListTemplatesByGroupOutput, error) {
	group, err := h.adminUC.GetTemplateGroup(ctx, input.GroupID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	resp := make([]dto.TemplateItem, len(group.Templates))
	for i, t := range group.Templates {
		resp[i] = dto.ToTemplateItem(t)
	}
	return &dto.ListTemplatesByGroupOutput{Body: resp}, nil
}

// --- Interactive Forms ---

func (h *LibraryAdminHandler) HandleCreateInteractiveForm(ctx context.Context, input *dto.CreateInteractiveFormInput) (*dto.CreateInteractiveFormOutput, error) {
	form, err := h.adminUC.CreateInteractiveForm(ctx, usecase.CreateInteractiveFormInput{
		TemplateID:  input.TemplateID,
		Name:        input.Body.Name,
		Description: input.Body.Description,
		FormLayout:  input.Body.FormLayout,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.CreateInteractiveFormOutput{Body: struct {
		ID uuid.UUID `json:"id" doc:"Created form ID"`
	}{ID: form.ID}}, nil
}

func (h *LibraryAdminHandler) HandleGetInteractiveForm(ctx context.Context, input *dto.GetInteractiveFormInput) (*dto.GetInteractiveFormOutput, error) {
	form, err := h.adminUC.GetInteractiveFormByTemplate(ctx, input.TemplateID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.GetInteractiveFormOutput{Body: dto.ToInteractiveFormDetailResponse(form)}, nil
}

func (h *LibraryAdminHandler) HandleUpdateInteractiveForm(ctx context.Context, input *dto.UpdateInteractiveFormInput) (*dto.UpdateInteractiveFormOutput, error) {
	form, err := h.adminUC.UpdateInteractiveForm(ctx, input.ID, usecase.UpdateInteractiveFormInput{
		Name:        input.Body.Name,
		Description: input.Body.Description,
		FormLayout:  input.Body.FormLayout,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UpdateInteractiveFormOutput{Body: dto.ToInteractiveFormDetailResponse(form)}, nil
}

func (h *LibraryAdminHandler) HandleDeleteInteractiveForm(ctx context.Context, input *dto.DeleteInteractiveFormInput) (*dto.DeleteInteractiveFormOutput, error) {
	if err := h.adminUC.DeleteInteractiveForm(ctx, input.ID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.DeleteInteractiveFormOutput{Body: struct {
		Message string `json:"message" doc:"Success message"`
	}{Message: i18n.T(ctx, "library.successes.interactiveFormDeleted")}}, nil
}

// --- Download Logs ---

func (h *LibraryAdminHandler) HandleGetDownloadLogs(ctx context.Context, input *dto.ListDownloadLogsInput) (*dto.ListDownloadLogsOutput, error) {
	q := dto.ToQueryOptions(input.Page, input.PageSize)
	var groupID *uuid.UUID
	if input.GroupID != "" {
		id, err := uuid.Parse(input.GroupID)
		if err != nil {
			return nil, apperrors.ToHumaError(ctx, apperrors.InvalidInputError("groupId", "library.errors.invalidFileType"))
		}
		groupID = &id
	}
	logs, err := h.adminUC.GetDownloadLogs(ctx, groupID, q)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	total := int64(0)
	if len(logs) > 0 {
		var allLogs []*entity.LibraryTemplateDownload
		allLogs, err = h.adminUC.GetDownloadLogs(ctx, nil, query.QueryOptions{Page: 1, PageSize: 100000})
		if err == nil {
			total = int64(len(allLogs))
		}
	}

	data := make([]dto.DownloadLogResponse, len(logs))
	for i, l := range logs {
		tmplTitle := ""
		groupName := ""
		if tmpl, err := h.adminUC.GetTemplate(ctx, l.TemplateID); err == nil && tmpl != nil {
			tmplTitle = tmpl.Title
		}
		if grp, err := h.adminUC.GetTemplateGroup(ctx, l.GroupID); err == nil && grp != nil {
			groupName = grp.Name
		}
		data[i] = dto.ToDownloadLogResponse(l, tmplTitle, groupName)
	}
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	return &dto.ListDownloadLogsOutput{Body: dto.DownloadLogListResponse{
		Data:       data,
		Total:      total,
		Page:       q.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}}, nil
}
