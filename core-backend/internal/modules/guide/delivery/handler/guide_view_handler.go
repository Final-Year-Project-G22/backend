package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/i18n"
)

type GuideViewHandler struct {
	guideViewUC usecase.GuideViewUseCase
}

func NewGuideViewHandler(guideViewUC usecase.GuideViewUseCase) *GuideViewHandler {
	return &GuideViewHandler{guideViewUC: guideViewUC}
}

func (h *GuideViewHandler) HandleListGuides(ctx context.Context, input *dto.ListGuidesInput) (*dto.ListGuidesOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))

	q := dto.ToQueryOptions(input.Page, input.PageSize)
	sectorIDs := dto.ParseCSVToUUIDs(input.SectorIDs)
	tagIDs := dto.ParseCSVToUUIDs(input.TagIDs)
	cards, err := h.guideViewUC.ListGuides(ctx, accountID, userID, q, input.Locale, sectorIDs, tagIDs)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	guides := make([]*dto.GuideCardDTO, 0, len(cards))
	for _, card := range cards {
		guides = append(guides, dto.ToGuideCardDTO(card))
	}

	return &dto.ListGuidesOutput{
		Body: dto.ListGuidesResponseBody{Guides: guides},
	}, nil
}

func (h *GuideViewHandler) HandleSearchGuides(ctx context.Context, input *dto.SearchGuidesInput) (*dto.SearchGuidesOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))

	q := dto.ToQueryOptions(input.Page, input.PageSize)
	cards, err := h.guideViewUC.SearchGuides(ctx, accountID, userID, input.Keyword, q, input.Locale)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	guides := make([]*dto.GuideCardDTO, 0, len(cards))
	for _, card := range cards {
		guides = append(guides, dto.ToGuideCardDTO(card))
	}

	return &dto.SearchGuidesOutput{
		Body: dto.SearchGuidesResponseBody{Guides: guides},
	}, nil
}

func (h *GuideViewHandler) HandleGetInProgressGuides(ctx context.Context, input *dto.GetInProgressGuidesInput) (*dto.GetInProgressGuidesOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))

	guides, err := h.guideViewUC.GetInProgressGuides(ctx, accountID, userID, input.Locale)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	result := make([]*dto.GuideWithProgressDTO, 0, len(guides))
	for _, g := range guides {
		result = append(result, dto.ToGuideWithProgressDTO(g))
	}

	return &dto.GetInProgressGuidesOutput{
		Body: dto.GetInProgressGuidesResponseBody{Guides: result},
	}, nil
}

func (h *GuideViewHandler) HandleGetCompletionStats(ctx context.Context, input *dto.GetCompletionStatsInput) (*dto.GetCompletionStatsOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))

	stats, err := h.guideViewUC.GetCompletionStats(ctx, accountID, userID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.GetCompletionStatsOutput{
		Body: *dto.ToCompletionStatsDTO(stats),
	}, nil
}

func (h *GuideViewHandler) HandleGetRecentlyViewed(ctx context.Context, input *dto.GetRecentlyViewedInput) (*dto.GetRecentlyViewedOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))

	q := dto.ToQueryOptions(input.Page, input.PageSize)
	cards, err := h.guideViewUC.GetRecentlyViewed(ctx, accountID, userID, q, input.Locale)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	guides := make([]*dto.GuideCardDTO, 0, len(cards))
	for _, card := range cards {
		guides = append(guides, dto.ToGuideCardDTO(card))
	}

	return &dto.GetRecentlyViewedOutput{
		Body: dto.GetRecentlyViewedResponseBody{Guides: guides},
	}, nil
}

func (h *GuideViewHandler) HandleGetPersonalizedGuide(ctx context.Context, input *dto.GetPersonalizedGuideInput) (*dto.GetPersonalizedGuideOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))

	guide, err := h.guideViewUC.GetPersonalizedGuide(ctx, accountID, userID, input.GuideSlug, input.Locale)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	steps := make([]*dto.PersonalizedStepDTO, 0, len(guide.Steps))
	for _, step := range guide.Steps {
		steps = append(steps, dto.ToPersonalizedStepDTO(step))
	}

	return &dto.GetPersonalizedGuideOutput{
		Body: dto.GetPersonalizedGuideResponseBody{
			ID:          guide.ID,
			Slug:        guide.Slug,
			Name:        guide.Name,
			Description: guide.Description,
			Steps:       steps,
			Progress:    dto.ToGuideProgressSummaryDTO(guide.Progress),
		},
	}, nil
}

func (h *GuideViewHandler) HandleGetCurrentStep(ctx context.Context, input *dto.GetCurrentStepInput) (*dto.GetCurrentStepOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))

	step, err := h.guideViewUC.GetCurrentStep(ctx, accountID, userID, input.GuideSlug, input.Locale)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	if step == nil {
		return &dto.GetCurrentStepOutput{
			Body: dto.GetCurrentStepResponseBody{},
		}, nil
	}

	return &dto.GetCurrentStepOutput{
		Body: dto.GetCurrentStepResponseBody{
			ID:            step.ID,
			Slug:          step.Slug,
			Title:         step.Title,
			Description:   step.Description,
			StepType:      step.StepType,
			SortOrder:     step.SortOrder,
			IsOptional:    step.IsOptional,
			EstimatedTime: step.EstimatedTime,
		},
	}, nil
}

func (h *GuideViewHandler) HandleStartStep(ctx context.Context, input *dto.StartStepInput) (*dto.StartStepOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))

	if err := h.guideViewUC.StartStep(ctx, accountID, userID, input.StepID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.StartStepOutput{
		Body: dto.StartStepResponseBody{Message: i18n.T(ctx, "guide.successes.stepStarted")},
	}, nil
}

func (h *GuideViewHandler) HandleCompleteStep(ctx context.Context, input *dto.CompleteStepInput) (*dto.CompleteStepOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))

	if err := h.guideViewUC.CompleteStep(ctx, accountID, userID, input.StepID, usecase.CompleteStepInput{
		UploadedDocuments: input.Body.UploadedDocuments,
		Notes:             input.Body.Notes,
		TimeSpentSeconds:  input.Body.TimeSpentSeconds,
	}); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.CompleteStepOutput{
		Body: dto.CompleteStepResponseBody{Message: i18n.T(ctx, "guide.successes.stepCompleted")},
	}, nil
}

func (h *GuideViewHandler) HandleMarkStepIncomplete(ctx context.Context, input *dto.MarkStepIncompleteInput) (*dto.MarkStepIncompleteOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))

	if err := h.guideViewUC.MarkStepIncomplete(ctx, accountID, userID, input.StepID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.MarkStepIncompleteOutput{
		Body: dto.MarkStepIncompleteResponseBody{Message: i18n.T(ctx, "guide.successes.stepMarkedIncomplete")},
	}, nil
}

func (h *GuideViewHandler) HandleSkipOptionalStep(ctx context.Context, input *dto.SkipOptionalStepInput) (*dto.SkipOptionalStepOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))

	if err := h.guideViewUC.SkipOptionalStep(ctx, accountID, userID, input.StepID, input.Reason); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.SkipOptionalStepOutput{
		Body: dto.SkipOptionalStepResponseBody{Message: i18n.T(ctx, "guide.successes.stepSkipped")},
	}, nil
}

func (h *GuideViewHandler) HandleUpdateProgress(ctx context.Context, input *dto.UpdateProgressInput) (*dto.UpdateProgressOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))

	if err := h.guideViewUC.UpdateStepProgress(ctx, accountID, userID, input.StepID, usecase.UpdateProgressInput{
		UploadedDocuments: input.Body.UploadedDocuments,
		Notes:             input.Body.Notes,
		TimeSpentSeconds:  input.Body.TimeSpentSeconds,
	}); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.UpdateProgressOutput{
		Body: dto.UpdateProgressResponseBody{Message: i18n.T(ctx, "guide.successes.progressUpdated")},
	}, nil
}

func (h *GuideViewHandler) HandleAddBookmark(ctx context.Context, input *dto.AddBookmarkInput) (*dto.AddBookmarkOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))

	if err := h.guideViewUC.AddBookmark(ctx, accountID, userID, input.StepID, input.Note); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.AddBookmarkOutput{
		Body: dto.AddBookmarkResponseBody{Message: i18n.T(ctx, "guide.successes.bookmarkAdded")},
	}, nil
}

func (h *GuideViewHandler) HandleUpdateBookmarkNote(ctx context.Context, input *dto.UpdateBookmarkNoteInput) (*dto.UpdateBookmarkNoteOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))

	if err := h.guideViewUC.UpdateBookmarkNote(ctx, accountID, userID, input.StepID, input.Body.Note); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.UpdateBookmarkNoteOutput{
		Body: dto.UpdateBookmarkNoteResponseBody{Message: i18n.T(ctx, "guide.successes.bookmarkUpdated")},
	}, nil
}

func (h *GuideViewHandler) HandleRemoveBookmark(ctx context.Context, input *dto.RemoveBookmarkInput) (*dto.RemoveBookmarkOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))

	if err := h.guideViewUC.RemoveBookmark(ctx, accountID, userID, input.StepID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.RemoveBookmarkOutput{
		Body: dto.RemoveBookmarkResponseBody{Message: i18n.T(ctx, "guide.successes.bookmarkRemoved")},
	}, nil
}

func (h *GuideViewHandler) HandleListBookmarks(ctx context.Context, input *dto.ListBookmarksInput) (*dto.ListBookmarksOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))

	q := dto.ToQueryOptions(input.Page, input.PageSize)
	bookmarks, err := h.guideViewUC.ListBookmarks(ctx, accountID, userID, q)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	result := make([]*dto.BookmarkWithStepDTO, 0, len(bookmarks))
	for _, b := range bookmarks {
		result = append(result, dto.ToBookmarkWithStepDTO(b))
	}

	return &dto.ListBookmarksOutput{
		Body: dto.ListBookmarksResponseBody{Bookmarks: result},
	}, nil
}
