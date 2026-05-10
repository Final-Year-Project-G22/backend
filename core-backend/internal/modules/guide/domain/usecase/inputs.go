package usecase

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/entity"
	"github.com/google/uuid"
)

type CompleteStepInput struct {
	UploadedDocuments []string `json:"uploadedDocuments,omitempty"`
	Notes             *string  `json:"notes,omitempty"`
	TimeSpentSeconds  *int     `json:"timeSpentSeconds,omitempty"`
}

type UpdateProgressInput struct {
	UploadedDocuments []string `json:"uploadedDocuments,omitempty"`
	Notes             *string  `json:"notes,omitempty"`
	TimeSpentSeconds  *int     `json:"timeSpentSeconds,omitempty"`
}

type CreateGuideInput struct {
	SectorIDs    []uuid.UUID        `json:"sectorIds"`
	TagIDs       []uuid.UUID        `json:"tagIds"`
	Slug         string             `json:"slug"`
	Icon         *string            `json:"icon,omitempty"`
	SortOrder    int                `json:"sortOrder"`
	Translations []TranslationInput `json:"translations,omitempty"`
	Conditions   []ConditionInput   `json:"conditions,omitempty"`
}

type UpdateGuideInput struct {
	SectorIDs    []uuid.UUID        `json:"sectorIds,omitempty"`
	TagIDs       []uuid.UUID        `json:"tagIds,omitempty"`
	Slug         *string            `json:"slug,omitempty"`
	Icon         *string            `json:"icon,omitempty"`
	SortOrder    *int               `json:"sortOrder,omitempty"`
	Translations []TranslationInput `json:"translations,omitempty"`
	Conditions   []ConditionInput   `json:"conditions,omitempty"`
}

type CreateStepInput struct {
	GuideID         uuid.UUID              `json:"guideId"`
	Slug            string                 `json:"slug"`
	StepType        entity.StepType        `json:"stepType"`
	SortOrder       int                    `json:"sortOrder"`
	IsOptional      bool                   `json:"isOptional"`
	EstimatedTime   *int                   `json:"estimatedTime,omitempty"`
	DifficultyLevel *int                   `json:"difficultyLevel,omitempty"`
	FeeEstimate     *int                   `json:"feeEstimate,omitempty"`
	EffectiveDate   *time.Time             `json:"effectiveDate,omitempty"`
	ExpiryDate      *time.Time             `json:"expiryDate,omitempty"`
	Translations    []StepTranslationInput `json:"translations,omitempty"`
	Conditions      []ConditionInput       `json:"conditions,omitempty"`
	Dependencies    []DependencyInput      `json:"dependencies,omitempty"`
}

type UpdateStepInput struct {
	Slug            *string                `json:"slug,omitempty"`
	StepType        *entity.StepType       `json:"stepType,omitempty"`
	SortOrder       *int                   `json:"sortOrder,omitempty"`
	IsOptional      *bool                  `json:"isOptional,omitempty"`
	EstimatedTime   *int                   `json:"estimatedTime,omitempty"`
	DifficultyLevel *int                   `json:"difficultyLevel,omitempty"`
	FeeEstimate     *int                   `json:"feeEstimate,omitempty"`
	EffectiveDate   *time.Time             `json:"effectiveDate,omitempty"`
	ExpiryDate      *time.Time             `json:"expiryDate,omitempty"`
	Translations    []StepTranslationInput `json:"translations,omitempty"`
	Conditions      []ConditionInput       `json:"conditions,omitempty"`
	Dependencies    []DependencyInput      `json:"dependencies,omitempty"`
}

type ConditionInput struct {
	ConditionType  string      `json:"conditionType"`
	Operator       string      `json:"operator"`
	ConditionValue interface{} `json:"conditionValue"`
	IsInverse      bool        `json:"isInverse"`
}

type TranslationInput struct {
	Language    string  `json:"language"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type StepTranslationInput struct {
	Language        string      `json:"language"`
	Title           string      `json:"title"`
	Description     *string     `json:"description,omitempty"`
	DetailedContent interface{} `json:"detailedContent,omitempty"`
}

type DependencyInput struct {
	RequiredStepID uuid.UUID             `json:"requiredStepId"`
	DependencyType entity.DependencyType `json:"dependencyType"`
}
