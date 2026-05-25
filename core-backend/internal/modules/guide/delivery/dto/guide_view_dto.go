package dto

import (
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type ListGuidesInput struct {
	Page      int    `query:"page" doc:"Page number"`
	PageSize  int    `query:"pageSize" doc:"Items per page"`
	SectorIDs string `query:"sectorIds" doc:"Comma-separated sector IDs"`
	TagIDs    string `query:"tagIds" doc:"Comma-separated tag IDs"`
}

type ListGuidesOutput struct {
	Body ListGuidesResponseBody
}

type ListGuidesResponseBody struct {
	Guides []*GuideCardDTO `json:"guides"`
}

type GuideCardDTO struct {
	ID          uuid.UUID   `json:"id" doc:"Guide ID"`
	Slug        string      `json:"slug" doc:"Guide slug"`
	Name        string      `json:"name" doc:"Localized guide name"`
	Description *string     `json:"description,omitempty" doc:"Localized description"`
	Icon        *string     `json:"icon,omitempty" doc:"Icon identifier"`
	SectorIDs   []uuid.UUID `json:"sectorIds" doc:"Target sector IDs"`
	TagIDs      []uuid.UUID `json:"tagIds" doc:"Target tag IDs"`
}

type SearchGuidesInput struct {
	Keyword  string `query:"q" doc:"Search keyword" minLength:"1"`
	Page     int    `query:"page" doc:"Page number"`
	PageSize int    `query:"pageSize" doc:"Items per page"`
}

type SearchGuidesOutput struct {
	Body SearchGuidesResponseBody
}

type SearchGuidesResponseBody struct {
	Guides []*GuideCardDTO `json:"guides"`
}

type GuideWithProgressDTO struct {
	ID             uuid.UUID `json:"id" doc:"Guide ID"`
	Slug           string    `json:"slug" doc:"Guide slug"`
	Name           string    `json:"name" doc:"Localized guide name"`
	Icon           *string   `json:"icon,omitempty" doc:"Icon identifier"`
	CompletedSteps int       `json:"completedSteps" doc:"Completed steps count"`
	TotalSteps     int       `json:"totalSteps" doc:"Total steps count"`
}

type GetInProgressGuidesInput struct{}

type GetInProgressGuidesOutput struct {
	Body GetInProgressGuidesResponseBody
}

type GetInProgressGuidesResponseBody struct {
	Guides []*GuideWithProgressDTO `json:"guides"`
}

type CompletionStatsDTO struct {
	CompletedGuides     int    `json:"completedGuides" doc:"Number of fully completed guides"`
	InProgressGuides    int    `json:"inProgressGuides" doc:"Number of guides in progress"`
	TotalStepsCompleted int    `json:"totalStepsCompleted" doc:"Total steps completed across all guides"`
	TotalStepsAll       int    `json:"totalStepsAll" doc:"Total steps across all guides"`
	Period              string `json:"period" doc:"Stats period (e.g. monthly)"`
}

type GetCompletionStatsInput struct{}

type GetCompletionStatsOutput struct {
	Body CompletionStatsDTO
}

func ToGuideWithProgressDTO(g *usecase.GuideWithProgress) *GuideWithProgressDTO {
	if g == nil {
		return nil
	}
	return &GuideWithProgressDTO{
		ID:             g.ID,
		Slug:           g.Slug,
		Name:           g.Name,
		Icon:           g.Icon,
		CompletedSteps: g.CompletedSteps,
		TotalSteps:     g.TotalSteps,
	}
}

func ToCompletionStatsDTO(s *usecase.CompletionStats) *CompletionStatsDTO {
	if s == nil {
		return nil
	}
	return &CompletionStatsDTO{
		CompletedGuides:     s.CompletedGuides,
		InProgressGuides:    s.InProgressGuides,
		TotalStepsCompleted: s.TotalStepsCompleted,
		TotalStepsAll:       s.TotalStepsAll,
		Period:              s.Period,
	}
}

type GetRecentlyViewedInput struct {
	Page     int `query:"page" doc:"Page number"`
	PageSize int `query:"pageSize" doc:"Items per page"`
}

type GetRecentlyViewedOutput struct {
	Body GetRecentlyViewedResponseBody
}

type GetRecentlyViewedResponseBody struct {
	Guides []*GuideCardDTO `json:"guides"`
}

type GetPersonalizedGuideInput struct {
	GuideSlug string `path:"guideSlug" doc:"Guide slug" minLength:"1"`
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
	GuideSlug string `path:"guideSlug" doc:"Guide slug" minLength:"1"`
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
		SectorIDs:   card.SectorIDs,
		TagIDs:      card.TagIDs,
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

func ParseCSVToUUIDs(csv string) []uuid.UUID {
	if strings.TrimSpace(csv) == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	result := make([]uuid.UUID, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			if id, err := uuid.Parse(trimmed); err == nil {
				result = append(result, id)
			}
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
