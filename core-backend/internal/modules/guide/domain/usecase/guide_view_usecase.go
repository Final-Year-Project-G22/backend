package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type GuideViewUseCase interface {
	GetCategoryTree(ctx context.Context, accountID, userID uuid.UUID, locale constants.Locale) ([]*CategoryNode, error)
	SearchGuides(ctx context.Context, accountID, userID uuid.UUID, keyword string, q query.QueryOptions, locale constants.Locale) ([]*GuideCard, error)
	GetRecentlyViewed(ctx context.Context, accountID, userID uuid.UUID, q query.QueryOptions, locale constants.Locale) ([]*GuideCard, error)

	GetPersonalizedGuide(ctx context.Context, accountID, userID uuid.UUID, guideSlug string, locale constants.Locale) (*PersonalizedGuide, error)
	GetCurrentStep(ctx context.Context, accountID, userID uuid.UUID, guideSlug string, locale constants.Locale) (*GetCurrentStepResult, error)

	StartStep(ctx context.Context, accountID, userID, stepID uuid.UUID) error
	CompleteStep(ctx context.Context, accountID, userID, stepID uuid.UUID, input CompleteStepInput) error
	MarkStepIncomplete(ctx context.Context, accountID, userID, stepID uuid.UUID) error
	SkipOptionalStep(ctx context.Context, accountID, userID, stepID uuid.UUID, reason *string) error
	UpdateStepProgress(ctx context.Context, accountID, userID, stepID uuid.UUID, input UpdateProgressInput) error

	AddBookmark(ctx context.Context, accountID, userID, stepID uuid.UUID, note *string) error
	UpdateBookmarkNote(ctx context.Context, accountID, userID, stepID uuid.UUID, note *string) error
	RemoveBookmark(ctx context.Context, accountID, userID, stepID uuid.UUID) error
	ListBookmarks(ctx context.Context, accountID, userID uuid.UUID, q query.QueryOptions) ([]*BookmarkWithStep, error)
}

type CategoryNode struct {
	ID          uuid.UUID       `json:"id"`
	Slug        string          `json:"slug"`
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Icon        *string         `json:"icon,omitempty"`
	SortOrder   int             `json:"sortOrder"`
	Children    []*CategoryNode `json:"children,omitempty"`
	Guides      []*GuideCard    `json:"guides,omitempty"`
}

type GuideCard struct {
	ID          uuid.UUID `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Icon        *string   `json:"icon,omitempty"`
	CategoryID  uuid.UUID `json:"categoryId"`
}

type PersonalizedGuide struct {
	ID          uuid.UUID             `json:"id"`
	Slug        string                `json:"slug"`
	Name        string                `json:"name"`
	Description *string               `json:"description,omitempty"`
	Steps       []*PersonalizedStep   `json:"steps"`
	Progress    *GuideProgressSummary `json:"progress,omitempty"`
}

type PersonalizedStep struct {
	ID            uuid.UUID             `json:"id"`
	Slug          string                `json:"slug"`
	Title         string                `json:"title"`
	Description   *string               `json:"description,omitempty"`
	StepType      entity.StepType       `json:"stepType"`
	SortOrder     int                   `json:"sortOrder"`
	IsOptional    bool                  `json:"isOptional"`
	Status        entity.ProgressStatus `json:"status"`
	EstimatedTime *int                  `json:"estimatedTime,omitempty"`
}

type GuideProgressSummary struct {
	TotalSteps      int `json:"totalSteps"`
	CompletedSteps  int `json:"completedSteps"`
	SkippedSteps    int `json:"skippedSteps"`
	InProgressSteps int `json:"inProgressSteps"`
}

type BookmarkWithStep struct {
	ID        uuid.UUID `json:"id"`
	StepID    uuid.UUID `json:"stepId"`
	Note      *string   `json:"note,omitempty"`
	StepTitle string    `json:"stepTitle"`
	GuideName string    `json:"guideName"`
	CreatedAt string    `json:"createdAt"`
}

type GetCurrentStepResult struct {
	ID            uuid.UUID       `json:"id"`
	Slug          string          `json:"slug"`
	Title         string          `json:"title"`
	Description   *string         `json:"description,omitempty"`
	StepType      entity.StepType `json:"stepType"`
	SortOrder     int             `json:"sortOrder"`
	IsOptional    bool            `json:"isOptional"`
	EstimatedTime *int            `json:"estimatedTime,omitempty"`
}
