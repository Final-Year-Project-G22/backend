package dto

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type GetCategoryTreeInput struct {
	Locale constants.Locale `query:"locale" doc:"Language locale (en, am)"`
}

type GetCategoryTreeOutput struct {
	Body GetCategoryTreeResponseBody
}

type GetCategoryTreeResponseBody struct {
	Categories []*CategoryNodeDTO `json:"categories"`
}

type CategoryNodeDTO struct {
	ID          uuid.UUID          `json:"id" doc:"Category ID"`
	Slug        string             `json:"slug" doc:"Category slug"`
	Name        string             `json:"name" doc:"Localized category name"`
	Description *string            `json:"description,omitempty" doc:"Localized description"`
	Icon        *string            `json:"icon,omitempty" doc:"Icon identifier"`
	SortOrder   int                `json:"sortOrder" doc:"Display order"`
	Children    []*CategoryNodeDTO `json:"children,omitempty" doc:"Nested subcategories"`
	Guides      []*GuideCardDTO    `json:"guides,omitempty" doc:"Guides in this category"`
}

type GuideCardDTO struct {
	ID          uuid.UUID `json:"id" doc:"Guide ID"`
	Slug        string    `json:"slug" doc:"Guide slug"`
	Name        string    `json:"name" doc:"Localized guide name"`
	Description *string   `json:"description,omitempty" doc:"Localized description"`
	Icon        *string   `json:"icon,omitempty" doc:"Icon identifier"`
	CategoryID  uuid.UUID `json:"categoryId" doc:"Parent category ID"`
}

type SearchGuidesInput struct {
	Keyword  string           `query:"q" doc:"Search keyword" minLength:"1"`
	Page     int              `query:"page" doc:"Page number"`
	PageSize int              `query:"pageSize" doc:"Items per page"`
	Locale   constants.Locale `query:"locale" doc:"Language locale (en, am)"`
}

type SearchGuidesOutput struct {
	Body SearchGuidesResponseBody
}

type SearchGuidesResponseBody struct {
	Guides []*GuideCardDTO `json:"guides"`
}

type GetRecentlyViewedInput struct {
	Page     int              `query:"page" doc:"Page number"`
	PageSize int              `query:"pageSize" doc:"Items per page"`
	Locale   constants.Locale `query:"locale" doc:"Language locale (en, am)"`
}

type GetRecentlyViewedOutput struct {
	Body GetRecentlyViewedResponseBody
}

type GetRecentlyViewedResponseBody struct {
	Guides []*GuideCardDTO `json:"guides"`
}

type GetPersonalizedGuideInput struct {
	GuideSlug string           `path:"guideSlug" doc:"Guide slug" minLength:"1"`
	Locale    constants.Locale `query:"locale" doc:"Language locale (en, am)"`
}

type GetPersonalizedGuideOutput struct {
	Body GetPersonalizedGuideResponseBody
}

type GetPersonalizedGuideResponseBody struct {
	ID          uuid.UUID                `json:"id" doc:"Guide ID"`
	Slug        string                   `json:"slug" doc:"Guide slug"`
	Name        string                   `json:"name" doc:"Localized guide name"`
	Description *string                  `json:"description,omitempty" doc:"Localized description"`
	Steps       []*PersonalizedStepDTO   `json:"steps" doc:"Guide steps with user progress"`
	Progress    *GuideProgressSummaryDTO `json:"progress,omitempty" doc:"User progress summary"`
}

type PersonalizedStepDTO struct {
	ID            uuid.UUID             `json:"id" doc:"Step ID"`
	Slug          string                `json:"slug" doc:"Step slug"`
	Title         string                `json:"title" doc:"Localized step title"`
	Description   *string               `json:"description,omitempty" doc:"Localized description"`
	StepType      entity.StepType       `json:"stepType" doc:"Step type"`
	SortOrder     int                   `json:"sortOrder" doc:"Display order"`
	IsOptional    bool                  `json:"isOptional" doc:"Whether step can be skipped"`
	Status        entity.ProgressStatus `json:"status" doc:"User progress status"`
	EstimatedTime *int                  `json:"estimatedTime,omitempty" doc:"Estimated time in minutes"`
}

type GuideProgressSummaryDTO struct {
	TotalSteps      int `json:"totalSteps" doc:"Total number of steps"`
	CompletedSteps  int `json:"completedSteps" doc:"Completed steps count"`
	SkippedSteps    int `json:"skippedSteps" doc:"Skipped steps count"`
	InProgressSteps int `json:"inProgressSteps" doc:"In-progress steps count"`
}

type GetCurrentStepInput struct {
	GuideSlug string           `path:"guideSlug" doc:"Guide slug" minLength:"1"`
	Locale    constants.Locale `query:"locale" doc:"Language locale (en, am)"`
}

type GetCurrentStepOutput struct {
	Body GetCurrentStepResponseBody
}

type GetCurrentStepResponseBody struct {
	ID            uuid.UUID       `json:"id" doc:"Step ID"`
	Slug          string          `json:"slug" doc:"Step slug"`
	Title         string          `json:"title" doc:"Localized step title"`
	Description   *string         `json:"description,omitempty" doc:"Localized description"`
	StepType      entity.StepType `json:"stepType" doc:"Step type"`
	SortOrder     int             `json:"sortOrder" doc:"Display order"`
	IsOptional    bool            `json:"isOptional" doc:"Whether step can be skipped"`
	EstimatedTime *int            `json:"estimatedTime,omitempty" doc:"Estimated time in minutes"`
}

type StartStepInput struct {
	StepID uuid.UUID `path:"stepId" doc:"Step ID"`
}

type StartStepOutput struct {
	Body StartStepResponseBody
}

type StartStepResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type CompleteStepRequest struct {
	UploadedDocuments []string `json:"uploadedDocuments,omitempty" doc:"Uploaded document URLs"`
	Notes             *string  `json:"notes,omitempty" doc:"Step notes"`
	TimeSpentSeconds  *int     `json:"timeSpentSeconds,omitempty" doc:"Time spent in seconds"`
}

type CompleteStepInput struct {
	StepID uuid.UUID `path:"stepId" doc:"Step ID"`
	Body   CompleteStepRequest
}

type CompleteStepOutput struct {
	Body CompleteStepResponseBody
}

type CompleteStepResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type MarkStepIncompleteInput struct {
	StepID uuid.UUID `path:"stepId" doc:"Step ID"`
}

type MarkStepIncompleteOutput struct {
	Body MarkStepIncompleteResponseBody
}

type MarkStepIncompleteResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type SkipOptionalStepInput struct {
	StepID uuid.UUID `path:"stepId" doc:"Step ID"`
	Reason *string   `json:"reason,omitempty" doc:"Reason for skipping"`
}

type SkipOptionalStepOutput struct {
	Body SkipOptionalStepResponseBody
}

type SkipOptionalStepResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type UpdateProgressRequest struct {
	UploadedDocuments []string `json:"uploadedDocuments,omitempty" doc:"Uploaded document URLs"`
	Notes             *string  `json:"notes,omitempty" doc:"Step notes"`
	TimeSpentSeconds  *int     `json:"timeSpentSeconds,omitempty" doc:"Time spent in seconds"`
}

type UpdateProgressInput struct {
	StepID uuid.UUID `path:"stepId" doc:"Step ID"`
	Body   UpdateProgressRequest
}

type UpdateProgressOutput struct {
	Body UpdateProgressResponseBody
}

type UpdateProgressResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type AddBookmarkInput struct {
	StepID uuid.UUID `path:"stepId" doc:"Step ID"`
	Note   *string   `json:"note,omitempty" doc:"Bookmark note" maxLength:"500"`
}

type AddBookmarkOutput struct {
	Body AddBookmarkResponseBody
}

type AddBookmarkResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type UpdateBookmarkNoteRequest struct {
	Note *string `json:"note" doc:"Bookmark note" maxLength:"500"`
}

type UpdateBookmarkNoteInput struct {
	StepID uuid.UUID `path:"stepId" doc:"Step ID"`
	Body   UpdateBookmarkNoteRequest
}

type UpdateBookmarkNoteOutput struct {
	Body UpdateBookmarkNoteResponseBody
}

type UpdateBookmarkNoteResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type RemoveBookmarkInput struct {
	StepID uuid.UUID `path:"stepId" doc:"Step ID"`
}

type RemoveBookmarkOutput struct {
	Body RemoveBookmarkResponseBody
}

type RemoveBookmarkResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type ListBookmarksInput struct {
	Page     int `query:"page" doc:"Page number"`
	PageSize int `query:"pageSize" doc:"Items per page"`
}

type ListBookmarksOutput struct {
	Body ListBookmarksResponseBody
}

type ListBookmarksResponseBody struct {
	Bookmarks []*BookmarkWithStepDTO `json:"bookmarks"`
}

type BookmarkWithStepDTO struct {
	ID        uuid.UUID `json:"id" doc:"Bookmark ID"`
	StepID    uuid.UUID `json:"stepId" doc:"Step ID"`
	Note      *string   `json:"note,omitempty" doc:"Bookmark note"`
	StepTitle string    `json:"stepTitle" doc:"Step title"`
	GuideName string    `json:"guideName" doc:"Guide name"`
	CreatedAt string    `json:"createdAt" doc:"Creation timestamp"`
}

func ToCategoryNodeDTO(node *usecase.CategoryNode) *CategoryNodeDTO {
	if node == nil {
		return nil
	}
	children := make([]*CategoryNodeDTO, 0, len(node.Children))
	for _, child := range node.Children {
		children = append(children, ToCategoryNodeDTO(child))
	}
	guides := make([]*GuideCardDTO, 0, len(node.Guides))
	for _, guide := range node.Guides {
		guides = append(guides, ToGuideCardDTO(guide))
	}
	return &CategoryNodeDTO{
		ID:          node.ID,
		Slug:        node.Slug,
		Name:        node.Name,
		Description: node.Description,
		Icon:        node.Icon,
		SortOrder:   node.SortOrder,
		Children:    children,
		Guides:      guides,
	}
}

func ToGuideCardDTO(card *usecase.GuideCard) *GuideCardDTO {
	if card == nil {
		return nil
	}
	return &GuideCardDTO{
		ID:          card.ID,
		Slug:        card.Slug,
		Name:        card.Name,
		Description: card.Description,
		Icon:        card.Icon,
		CategoryID:  card.CategoryID,
	}
}

func ToPersonalizedStepDTO(step *usecase.PersonalizedStep) *PersonalizedStepDTO {
	if step == nil {
		return nil
	}
	return &PersonalizedStepDTO{
		ID:            step.ID,
		Slug:          step.Slug,
		Title:         step.Title,
		Description:   step.Description,
		StepType:      step.StepType,
		SortOrder:     step.SortOrder,
		IsOptional:    step.IsOptional,
		Status:        step.Status,
		EstimatedTime: step.EstimatedTime,
	}
}

func ToGuideProgressSummaryDTO(summary *usecase.GuideProgressSummary) *GuideProgressSummaryDTO {
	if summary == nil {
		return nil
	}
	return &GuideProgressSummaryDTO{
		TotalSteps:      summary.TotalSteps,
		CompletedSteps:  summary.CompletedSteps,
		SkippedSteps:    summary.SkippedSteps,
		InProgressSteps: summary.InProgressSteps,
	}
}

func ToBookmarkWithStepDTO(b *usecase.BookmarkWithStep) *BookmarkWithStepDTO {
	if b == nil {
		return nil
	}
	return &BookmarkWithStepDTO{
		ID:        b.ID,
		StepID:    b.StepID,
		Note:      b.Note,
		StepTitle: b.StepTitle,
		GuideName: b.GuideName,
		CreatedAt: b.CreatedAt,
	}
}

func ToQueryOptions(page, pageSize int) query.QueryOptions {
	opts := query.QueryOptions{}
	if page > 0 {
		opts.Page = page
	}
	if pageSize > 0 {
		opts.PageSize = pageSize
	}
	return opts
}
