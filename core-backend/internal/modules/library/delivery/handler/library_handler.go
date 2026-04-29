package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
)

type LibraryHandler struct {
	viewUC usecase.LibraryViewUsecase
}

func NewLibraryHandler(viewUC usecase.LibraryViewUsecase) *LibraryHandler {
	return &LibraryHandler{viewUC: viewUC}
}

func (h *LibraryHandler) HandleListCategories(ctx context.Context, input *dto.LibraryListCategoriesInput) (*dto.LibraryListCategoriesOutput, error) {
	var locale *string
	if input.Locale != "" {
		locale = &input.Locale
	}
	categories, err := h.viewUC.ListCategories(ctx, locale)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	roots := buildCategoryTree(categories, nil)
	return &dto.LibraryListCategoriesOutput{Body: roots}, nil
}

func (h *LibraryHandler) HandleListTemplateGroups(ctx context.Context, input *dto.ListTemplateGroupsInput) (*dto.ListTemplateGroupsOutput, error) {
	q := dto.ToQueryOptions(input.Page, input.PageSize)
	q.Search = input.Search

	var categoryID *uuid.UUID
	if input.CategoryID != "" {
		id, err := uuid.Parse(input.CategoryID)
		if err != nil {
			return nil, apperrors.ToHumaError(ctx, apperrors.InvalidInputError("categoryId", "library.errors.invalidFileType"))
		}
		categoryID = &id
	}
	var format *entity.TemplateFormat
	if input.Format != "" {
		f := entity.TemplateFormat(input.Format)
		format = &f
	}

	groups, err := h.viewUC.ListTemplateGroups(ctx, categoryID, format, q)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	data := make([]dto.TemplateGroupCardResponse, len(groups))
	for i, g := range groups {
		data[i] = dto.ToTemplateGroupCardResponse(g)
	}

	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	total := int64(len(data))
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &dto.ListTemplateGroupsOutput{Body: dto.ListTemplateGroupsResponseBody{
		Data:       data,
		Total:      total,
		Page:       q.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}}, nil
}

func (h *LibraryHandler) HandleGetTemplateGroup(ctx context.Context, input *dto.GetTemplateGroupBySlugInput) (*dto.GetTemplateGroupDetailOutput, error) {
	var locale *string
	if input.Locale != "" {
		locale = &input.Locale
	}
	group, templates, err := h.viewUC.GetTemplateGroup(ctx, input.Slug, locale)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.GetTemplateGroupDetailOutput{Body: dto.ToGroupDetailResponse(group, templates)}, nil
}

func (h *LibraryHandler) HandleDownloadTemplate(ctx context.Context, input *dto.DownloadTemplateInput) (*dto.DownloadTemplateOutput, error) {
	var accountID *uuid.UUID
	if id := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID)); id != uuid.Nil {
		accountID = &id
	}
	var language *string
	if input.Language != "" {
		language = &input.Language
	}

	result, err := h.viewUC.DownloadTemplate(ctx, usecase.DownloadInput{
		Slug:      input.Slug,
		Language:  language,
		AccountID: accountID,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.DownloadTemplateOutput{Body: dto.DownloadTemplateResponseBody{
		PresignedURL: result.PresignedURL,
		ExpiresAt:    result.ExpiresAt,
		Filename:     result.Filename,
	}}, nil
}

func (h *LibraryHandler) HandleListMyDownloads(ctx context.Context, input *dto.ListMyDownloadsInput) (*dto.ListMyDownloadsOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if accountID == uuid.Nil {
		return nil, apperrors.ToHumaError(ctx, apperrors.UnauthorizedError("library.errors.authRequired"))
	}

	q := dto.ToQueryOptions(input.Page, input.PageSize)
	logs, err := h.viewUC.ListMyDownloads(ctx, accountID, q)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	data := make([]dto.MyDownloadResponse, len(logs))
	for i, d := range logs {
		data[i] = dto.ToMyDownloadResponse(d)
	}

	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	total := int64(len(data))
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &dto.ListMyDownloadsOutput{Body: dto.ListMyDownloadsResponseBody{
		Data:       data,
		Total:      total,
		Page:       q.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}}, nil
}

func buildCategoryTree(categories []*entity.LibraryCategory, parentID *uuid.UUID) []dto.CategoryNodeResponse {
	var nodes []dto.CategoryNodeResponse
	for _, cat := range categories {
		if (parentID == nil && cat.ParentCategoryID == nil) ||
			(parentID != nil && cat.ParentCategoryID != nil && *cat.ParentCategoryID == *parentID) {
			children := buildCategoryTree(categories, &cat.ID)
			nodes = append(nodes, dto.ToCategoryNodeResponse(cat, children))
		}
	}
	return nodes
}
