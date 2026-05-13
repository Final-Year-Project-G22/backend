package dto

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// --- Category ---

type CreateCategoryRequest struct {
	Slug         string                      `json:"slug" doc:"Category slug" minLength:"1" maxLength:"100"`
	Icon         *string                     `json:"icon,omitempty" doc:"Icon identifier" maxLength:"50"`
	SortOrder    int                         `json:"sortOrder" doc:"Display order"`
	ParentID     *uuid.UUID                  `json:"parentId,omitempty" doc:"Parent category ID"`
	Translations []CreateCategoryTranslation `json:"translations,omitempty" doc:"Localized translations"`
	Conditions   []CreateCategoryCondition   `json:"conditions,omitempty" doc:"Visibility conditions"`
}

type CreateCategoryTranslation struct {
	Language    string  `json:"language" doc:"Language code (en, am)" minLength:"2" maxLength:"5"`
	Name        string  `json:"name" doc:"Localized name" minLength:"1" maxLength:"200"`
	Description *string `json:"description,omitempty" doc:"Localized description"`
}

type CreateCategoryCondition struct {
	ConditionType  string      `json:"conditionType" doc:"Condition type" minLength:"1"`
	Operator       string      `json:"operator" doc:"Comparison operator" minLength:"1"`
	ConditionValue interface{} `json:"conditionValue" doc:"Condition value"`
	IsInverse      bool        `json:"isInverse" doc:"Whether to invert the condition"`
}

type CreateCategoryInput struct {
	Body CreateCategoryRequest
}

type CreateCategoryOutput struct {
	Body CreateCategoryResponseBody
}

type CreateCategoryResponseBody struct {
	ID uuid.UUID `json:"id" doc:"Created category ID"`
}

type UpdateCategoryRequest struct {
	Slug         *string                     `json:"slug,omitempty" doc:"Category slug" maxLength:"100"`
	Icon         *string                     `json:"icon,omitempty" doc:"Icon identifier" maxLength:"50"`
	SortOrder    *int                        `json:"sortOrder,omitempty" doc:"Display order"`
	ParentID     *uuid.UUID                  `json:"parentId,omitempty" doc:"Parent category ID"`
	Translations []UpdateCategoryTranslation `json:"translations,omitempty" doc:"Localized translations"`
	Conditions   []UpdateCategoryCondition   `json:"conditions,omitempty" doc:"Visibility conditions"`
}

type UpdateCategoryTranslation struct {
	Language    string  `json:"language" doc:"Language code (en, am)" minLength:"2" maxLength:"5"`
	Name        string  `json:"name" doc:"Localized name" minLength:"1" maxLength:"200"`
	Description *string `json:"description,omitempty" doc:"Localized description"`
}

type UpdateCategoryCondition struct {
	ConditionType  string      `json:"conditionType" doc:"Condition type" minLength:"1"`
	Operator       string      `json:"operator" doc:"Comparison operator" minLength:"1"`
	ConditionValue interface{} `json:"conditionValue" doc:"Condition value"`
	IsInverse      bool        `json:"isInverse" doc:"Whether to invert the condition"`
}

type UpdateCategoryInput struct {
	ID   uuid.UUID `path:"id" doc:"Category ID"`
	Body UpdateCategoryRequest
}

type UpdateCategoryOutput struct {
	Body UpdateCategoryResponseBody
}

type UpdateCategoryResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type DeleteCategoryInput struct {
	ID uuid.UUID `path:"id" doc:"Category ID"`
}

type DeleteCategoryOutput struct {
	Body DeleteCategoryResponseBody
}

type DeleteCategoryResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type AddCategoryConditionRequest struct {
	ConditionType  string      `json:"conditionType" doc:"Condition type" minLength:"1"`
	Operator       string      `json:"operator" doc:"Comparison operator" minLength:"1"`
	ConditionValue interface{} `json:"conditionValue" doc:"Condition value"`
	IsInverse      bool        `json:"isInverse" doc:"Whether to invert the condition"`
}

type AddCategoryConditionInput struct {
	CategoryID uuid.UUID `path:"id" doc:"Category ID"`
	Body       AddCategoryConditionRequest
}

type AddCategoryConditionOutput struct {
	Body AddCategoryConditionResponseBody
}

type AddCategoryConditionResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type RemoveCategoryConditionInput struct {
	CondID uuid.UUID `path:"condId" doc:"Condition ID"`
}

type RemoveCategoryConditionOutput struct {
	Body RemoveCategoryConditionResponseBody
}

type RemoveCategoryConditionResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type SetCategoryTranslationsRequest struct {
	Translations []UpdateCategoryTranslation `json:"translations" doc:"Full set of translations"`
}

type SetCategoryTranslationsInput struct {
	ID   uuid.UUID `path:"id" doc:"Category ID"`
	Body SetCategoryTranslationsRequest
}

type SetCategoryTranslationsOutput struct {
	Body SetCategoryTranslationsResponseBody
}

type SetCategoryTranslationsResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Guide ---

type CreateGuideRequest struct {
	SectorIDs    []uuid.UUID              `json:"sectorIds" doc:"Target sector IDs"`
	TagIDs       []uuid.UUID              `json:"tagIds" doc:"Target tag IDs"`
	Slug         string                   `json:"slug" doc:"Guide slug" minLength:"1" maxLength:"100"`
	Icon         *string                  `json:"icon,omitempty" doc:"Icon identifier" maxLength:"50"`
	ImageURL     *string                  `json:"imageUrl,omitempty" doc:"Guide cover image URL" maxLength:"500"`
	SortOrder    int                      `json:"sortOrder" doc:"Display order"`
	Translations []CreateGuideTranslation `json:"translations,omitempty" doc:"Localized translations"`
	Conditions   []CreateGuideCondition   `json:"conditions,omitempty" doc:"Visibility conditions"`
}

type CreateGuideTranslation struct {
	Language    string  `json:"language" doc:"Language code (en, am)" minLength:"2" maxLength:"5"`
	Name        string  `json:"name" doc:"Localized name" minLength:"1" maxLength:"200"`
	Description *string `json:"description,omitempty" doc:"Localized description"`
}

type CreateGuideCondition struct {
	ConditionType  string      `json:"conditionType" doc:"Condition type" minLength:"1"`
	Operator       string      `json:"operator" doc:"Comparison operator" minLength:"1"`
	ConditionValue interface{} `json:"conditionValue" doc:"Condition value"`
	IsInverse      bool        `json:"isInverse" doc:"Whether to invert the condition"`
}

type CreateGuideInput struct {
	Body CreateGuideRequest
}

type CreateGuideOutput struct {
	Body CreateGuideResponseBody
}

type CreateGuideResponseBody struct {
	ID uuid.UUID `json:"id" doc:"Created guide ID"`
}

type UpdateGuideRequest struct {
	SectorIDs       []uuid.UUID              `json:"sectorIds,omitempty" doc:"Target sector IDs"`
	TagIDs          []uuid.UUID              `json:"tagIds,omitempty" doc:"Target tag IDs"`
	Slug            *string                  `json:"slug,omitempty" doc:"Guide slug" maxLength:"100"`
	Icon            *string                  `json:"icon,omitempty" doc:"Icon identifier" maxLength:"50"`
	ImageURL        *string                  `json:"imageUrl,omitempty" doc:"Guide cover image URL" maxLength:"500"`
	SortOrder       *int                     `json:"sortOrder,omitempty" doc:"Display order"`
	Translations    []UpdateGuideTranslation `json:"translations,omitempty" doc:"Localized translations"`
	TranslationMode *string                  `json:"translationMode,omitempty" doc:"Translation mode: 'merge' to upsert without deleting, anything else or absent for full replacement" enum:"merge,replace"`
	Conditions      []UpdateGuideCondition   `json:"conditions,omitempty" doc:"Visibility conditions"`
}

type UpdateGuideTranslation struct {
	Language    string  `json:"language" doc:"Language code (en, am)" minLength:"2" maxLength:"5"`
	Name        string  `json:"name" doc:"Localized name" minLength:"1" maxLength:"200"`
	Description *string `json:"description,omitempty" doc:"Localized description"`
}

type UpdateGuideCondition struct {
	ConditionType  string      `json:"conditionType" doc:"Condition type" minLength:"1"`
	Operator       string      `json:"operator" doc:"Comparison operator" minLength:"1"`
	ConditionValue interface{} `json:"conditionValue" doc:"Condition value"`
	IsInverse      bool        `json:"isInverse" doc:"Whether to invert the condition"`
}

type UpdateGuideInput struct {
	ID   uuid.UUID `path:"id" doc:"Guide ID"`
	Body UpdateGuideRequest
}

type UpdateGuideOutput struct {
	Body UpdateGuideResponseBody
}

type UpdateGuideResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type DeleteGuideInput struct {
	ID uuid.UUID `path:"id" doc:"Guide ID"`
}

type DeleteGuideOutput struct {
	Body DeleteGuideResponseBody
}

type DeleteGuideResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type AddGuideConditionRequest struct {
	ConditionType  string      `json:"conditionType" doc:"Condition type" minLength:"1"`
	Operator       string      `json:"operator" doc:"Comparison operator" minLength:"1"`
	ConditionValue interface{} `json:"conditionValue" doc:"Condition value"`
	IsInverse      bool        `json:"isInverse" doc:"Whether to invert the condition"`
}

type AddGuideConditionInput struct {
	GuideID uuid.UUID `path:"id" doc:"Guide ID"`
	Body    AddGuideConditionRequest
}

type AddGuideConditionOutput struct {
	Body AddGuideConditionResponseBody
}

type AddGuideConditionResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type RemoveGuideConditionInput struct {
	CondID uuid.UUID `path:"condId" doc:"Condition ID"`
}

type RemoveGuideConditionOutput struct {
	Body RemoveGuideConditionResponseBody
}

type RemoveGuideConditionResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type SetGuideTranslationsRequest struct {
	Translations []UpdateGuideTranslation `json:"translations" doc:"Full set of translations"`
}

type SetGuideTranslationsInput struct {
	ID              uuid.UUID `path:"id" doc:"Guide ID"`
	TranslationMode string    `query:"translationMode" doc:"Translation mode: 'merge' to upsert without deleting, absent for full replacement" enum:"merge,replace"`
	Body            SetGuideTranslationsRequest
}

type SetGuideTranslationsOutput struct {
	Body SetGuideTranslationsResponseBody
}

type SetGuideTranslationsResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Guide Image ---

type UploadGuideImageFormData struct {
	File huma.FormFile `form:"file"`
}

type UploadGuideImageInput struct {
	ID      uuid.UUID `path:"id" doc:"Guide ID"`
	RawBody huma.MultipartFormFiles[UploadGuideImageFormData]
}

type UploadGuideImageResponseBody struct {
	ImageURL string `json:"imageUrl" doc:"Uploaded image URL"`
}

type UploadGuideImageOutput struct {
	Body UploadGuideImageResponseBody
}

// --- Step ---

type CreateStepRequest struct {
	GuideID         uuid.UUID               `json:"guideId" doc:"Parent guide ID"`
	Slug            string                  `json:"slug" doc:"Step slug" minLength:"1" maxLength:"100"`
	StepType        entity.StepType         `json:"stepType" doc:"Step type"`
	SortOrder       int                     `json:"sortOrder" doc:"Display order"`
	IsOptional      bool                    `json:"isOptional" doc:"Whether step can be skipped"`
	EstimatedTime   *int                    `json:"estimatedTime,omitempty" doc:"Estimated time in minutes"`
	DifficultyLevel *int                    `json:"difficultyLevel,omitempty" doc:"Difficulty level (1-5)"`
	FeeEstimate     *int                    `json:"feeEstimate,omitempty" doc:"Estimated fee"`
	EffectiveDate   *time.Time              `json:"effectiveDate,omitempty" doc:"When step becomes active"`
	ExpiryDate      *time.Time              `json:"expiryDate,omitempty" doc:"When step expires"`
	Translations    []CreateStepTranslation `json:"translations,omitempty" doc:"Localized translations"`
	Conditions      []CreateStepCondition   `json:"conditions,omitempty" doc:"Visibility conditions"`
	Dependencies    []CreateStepDependency  `json:"dependencies,omitempty" doc:"Step dependencies"`
}

type CreateStepTranslation struct {
	Language          string      `json:"language" doc:"Language code (en, am)" minLength:"2" maxLength:"5"`
	Title             string      `json:"title" doc:"Localized title" minLength:"1" maxLength:"200"`
	Description       *string     `json:"description,omitempty" doc:"Localized description"`
	DetailedContent   interface{} `json:"detailedContent,omitempty" doc:"Rich content as JSON"`
	RequiredDocuments interface{} `json:"requiredDocuments,omitempty" doc:"Required documents as JSON array"`
}

type CreateStepCondition struct {
	ConditionType  string      `json:"conditionType" doc:"Condition type" minLength:"1"`
	Operator       string      `json:"operator" doc:"Comparison operator" minLength:"1"`
	ConditionValue interface{} `json:"conditionValue" doc:"Condition value"`
	IsInverse      bool        `json:"isInverse" doc:"Whether to invert the condition"`
}

type CreateStepDependency struct {
	RequiredStepID uuid.UUID             `json:"requiredStepId" doc:"Required step ID"`
	DependencyType entity.DependencyType `json:"dependencyType" doc:"Dependency type"`
}

type CreateStepInput struct {
	Body CreateStepRequest
}

type CreateStepOutput struct {
	Body CreateStepResponseBody
}

type CreateStepResponseBody struct {
	ID uuid.UUID `json:"id" doc:"Created step ID"`
}

type UpdateStepRequest struct {
	Slug            *string                 `json:"slug,omitempty" doc:"Step slug" maxLength:"100"`
	StepType        *entity.StepType        `json:"stepType,omitempty" doc:"Step type"`
	SortOrder       *int                    `json:"sortOrder,omitempty" doc:"Display order"`
	IsOptional      *bool                   `json:"isOptional,omitempty" doc:"Whether step can be skipped"`
	EstimatedTime   *int                    `json:"estimatedTime,omitempty" doc:"Estimated time in minutes"`
	DifficultyLevel *int                    `json:"difficultyLevel,omitempty" doc:"Difficulty level (1-5)"`
	FeeEstimate     *int                    `json:"feeEstimate,omitempty" doc:"Estimated fee"`
	EffectiveDate   *time.Time              `json:"effectiveDate,omitempty" doc:"When step becomes active"`
	ExpiryDate      *time.Time              `json:"expiryDate,omitempty" doc:"When step expires"`
	Translations    []UpdateStepTranslation `json:"translations,omitempty" doc:"Localized translations"`
	TranslationMode *string                 `json:"translationMode,omitempty" doc:"Translation mode: 'merge' to upsert without deleting, anything else or absent for full replacement" enum:"merge,replace"`
	Conditions      []UpdateStepCondition   `json:"conditions,omitempty" doc:"Visibility conditions"`
	Dependencies    []UpdateStepDependency  `json:"dependencies,omitempty" doc:"Step dependencies"`
}

type UpdateStepTranslation struct {
	Language          string      `json:"language" doc:"Language code (en, am)" minLength:"2" maxLength:"5"`
	Title             string      `json:"title" doc:"Localized title" minLength:"1" maxLength:"200"`
	Description       *string     `json:"description,omitempty" doc:"Localized description"`
	DetailedContent   interface{} `json:"detailedContent,omitempty" doc:"Rich content as JSON"`
	RequiredDocuments interface{} `json:"requiredDocuments,omitempty" doc:"Required documents as JSON array"`
}

type UpdateStepCondition struct {
	ConditionType  string      `json:"conditionType" doc:"Condition type" minLength:"1"`
	Operator       string      `json:"operator" doc:"Comparison operator" minLength:"1"`
	ConditionValue interface{} `json:"conditionValue" doc:"Condition value"`
	IsInverse      bool        `json:"isInverse" doc:"Whether to invert the condition"`
}

type UpdateStepDependency struct {
	RequiredStepID uuid.UUID             `json:"requiredStepId" doc:"Required step ID"`
	DependencyType entity.DependencyType `json:"dependencyType" doc:"Dependency type"`
}

type UpdateStepInput struct {
	ID   uuid.UUID `path:"id" doc:"Step ID"`
	Body UpdateStepRequest
}

type UpdateStepOutput struct {
	Body UpdateStepResponseBody
}

type UpdateStepResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type DeleteStepInput struct {
	ID uuid.UUID `path:"id" doc:"Step ID"`
}

type DeleteStepOutput struct {
	Body DeleteStepResponseBody
}

type DeleteStepResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type ReorderStepsRequest struct {
	GuideID uuid.UUID   `json:"guideId" doc:"Guide ID"`
	StepIDs []uuid.UUID `json:"stepIds" doc:"Ordered list of step IDs"`
}

type ReorderStepsInput struct {
	Body ReorderStepsRequest
}

type ReorderStepsOutput struct {
	Body ReorderStepsResponseBody
}

type ReorderStepsResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type AddStepConditionRequest struct {
	ConditionType  string      `json:"conditionType" doc:"Condition type" minLength:"1"`
	Operator       string      `json:"operator" doc:"Comparison operator" minLength:"1"`
	ConditionValue interface{} `json:"conditionValue" doc:"Condition value"`
	IsInverse      bool        `json:"isInverse" doc:"Whether to invert the condition"`
}

type AddStepConditionInput struct {
	StepID uuid.UUID `path:"id" doc:"Step ID"`
	Body   AddStepConditionRequest
}

type AddStepConditionOutput struct {
	Body AddStepConditionResponseBody
}

type AddStepConditionResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type RemoveStepConditionInput struct {
	CondID uuid.UUID `path:"condId" doc:"Condition ID"`
}

type RemoveStepConditionOutput struct {
	Body RemoveStepConditionResponseBody
}

type RemoveStepConditionResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type AddStepDependencyRequest struct {
	RequiredStepID uuid.UUID             `json:"requiredStepId" doc:"Required step ID"`
	DependencyType entity.DependencyType `json:"dependencyType" doc:"Dependency type"`
}

type AddStepDependencyInput struct {
	StepID uuid.UUID `path:"id" doc:"Step ID"`
	Body   AddStepDependencyRequest
}

type AddStepDependencyOutput struct {
	Body AddStepDependencyResponseBody
}

type AddStepDependencyResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type RemoveStepDependencyInput struct {
	DepID uuid.UUID `path:"depId" doc:"Dependency ID"`
}

type RemoveStepDependencyOutput struct {
	Body RemoveStepDependencyResponseBody
}

type RemoveStepDependencyResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type SetStepTranslationsRequest struct {
	Translations []UpdateStepTranslation `json:"translations" doc:"Full set of translations"`
}

type SetStepTranslationsInput struct {
	ID              uuid.UUID `path:"id" doc:"Step ID"`
	TranslationMode string    `query:"translationMode" doc:"Translation mode: 'merge' to upsert without deleting, absent for full replacement" enum:"merge,replace"`
	Body            SetStepTranslationsRequest
}

type SetStepTranslationsOutput struct {
	Body SetStepTranslationsResponseBody
}

type SetStepTranslationsResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type GetStepVersionsInput struct {
	StepID   uuid.UUID `path:"id" doc:"Step ID"`
	Page     int       `query:"page" doc:"Page number"`
	PageSize int       `query:"pageSize" doc:"Items per page"`
}

type GetStepVersionsOutput struct {
	Body GetStepVersionsResponseBody
}

type GetStepVersionsResponseBody struct {
	Versions []StepVersionDTO `json:"versions"`
}

type StepVersionDTO struct {
	ID            uuid.UUID  `json:"id" doc:"Version record ID"`
	Version       int        `json:"version" doc:"Version number"`
	StepID        uuid.UUID  `json:"stepId" doc:"Step ID"`
	EffectiveDate time.Time  `json:"effectiveDate" doc:"When this version became active"`
	CreatedAt     *time.Time `json:"createdAt" doc:"When this version was created"`
}

type RevertStepToVersionInput struct {
	StepID  uuid.UUID `path:"id" doc:"Step ID"`
	Version int       `path:"version" doc:"Version number to revert to"`
}

type RevertStepToVersionOutput struct {
	Body RevertStepToVersionResponseBody
}

type RevertStepToVersionResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Journey ---

type InvalidateUserJourneyRequest struct {
	GuideID uuid.UUID `json:"guideId" doc:"Guide ID"`
}

type InvalidateUserJourneyInput struct {
	UserID uuid.UUID `path:"userId" doc:"User ID"`
	Body   InvalidateUserJourneyRequest
}

type InvalidateUserJourneyOutput struct {
	Body InvalidateUserJourneyResponseBody
}

type InvalidateUserJourneyResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type InvalidateAllJourneysInput struct {
	GuideID uuid.UUID `path:"guideId" doc:"Guide ID"`
}

type InvalidateAllJourneysOutput struct {
	Body InvalidateAllJourneysResponseBody
}

type InvalidateAllJourneysResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Admin Read Models ---

type AdminPaginationQuery struct {
	Page      int      `query:"page" doc:"Page number"`
	PageSize  int      `query:"pageSize" doc:"Items per page"`
	Search    string   `query:"search" doc:"Search keyword"`
	SortBy    []string `query:"sortBy" doc:"Sort columns"`
	SortOrder []string `query:"sortOrder" doc:"Sort order (asc, desc)"`
}

type GuideCategoryTreeAdminInput struct {
	IncludeInactive bool             `query:"includeInactive" doc:"Include inactive categories"`
	Locale          constants.Locale `query:"locale" doc:"Language locale (en, am)"`
}

type GuideCategoryTreeAdminOutput struct {
	Body GuideCategoryTreeAdminResponseBody
}

type GuideCategoryTreeAdminResponseBody struct {
	Categories []AdminCategoryDTO `json:"categories"`
}

type ListGuidesAdminInput struct {
	AdminPaginationQuery
	Locale constants.Locale `query:"locale" doc:"Language locale (en, am)"`
}

type ListGuidesAdminOutput struct {
	Body ListGuidesAdminResponseBody
}

type ListGuidesAdminResponseBody struct {
	Guides     []AdminGuideCardDTO `json:"guides"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"pageSize"`
	TotalItems int64               `json:"totalItems"`
	TotalPages int                 `json:"totalPages"`
}

type GetGuideAdminInput struct {
	ID     uuid.UUID        `path:"id" doc:"Guide ID"`
	Locale constants.Locale `query:"locale" doc:"Language locale (en, am)"`
}

type GetGuideAdminOutput struct {
	Body GetGuideAdminResponseBody
}

type GetGuideAdminResponseBody struct {
	Guide AdminGuideDetailDTO `json:"guide"`
}

type ListGuideStepsAdminInput struct {
	ID uuid.UUID `path:"id" doc:"Guide ID"`
	AdminPaginationQuery
	Locale constants.Locale `query:"locale" doc:"Language locale (en, am)"`
}

type ListGuideStepsAdminOutput struct {
	Body ListGuideStepsAdminResponseBody
}

type ListGuideStepsAdminResponseBody struct {
	Steps      []AdminGuideStepDTO `json:"steps"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"pageSize"`
	TotalItems int64               `json:"totalItems"`
	TotalPages int                 `json:"totalPages"`
}

type AdminCategoryDTO struct {
	ID          uuid.UUID          `json:"id"`
	ParentID    *uuid.UUID         `json:"parentId,omitempty"`
	Slug        string             `json:"slug"`
	Name        string             `json:"name"`
	Description *string            `json:"description,omitempty"`
	Icon        *string            `json:"icon,omitempty"`
	SortOrder   int                `json:"sortOrder"`
	Children    []AdminCategoryDTO `json:"children,omitempty"`
}

type AdminGuideCardDTO struct {
	ID          uuid.UUID   `json:"id"`
	Slug        string      `json:"slug"`
	Name        string      `json:"name"`
	Description *string     `json:"description,omitempty"`
	Icon        *string     `json:"icon,omitempty"`
	ImageURL    *string     `json:"imageUrl,omitempty"`
	SectorIDs   []uuid.UUID `json:"sectorIds"`
	TagIDs      []uuid.UUID `json:"tagIds"`
	SortOrder   int         `json:"sortOrder"`
}

type AdminTranslationDTO struct {
	Language    string  `json:"language"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type AdminConditionDTO struct {
	ID             uuid.UUID   `json:"id"`
	ConditionType  string      `json:"conditionType"`
	Operator       string      `json:"operator"`
	ConditionValue interface{} `json:"conditionValue"`
	IsInverse      bool        `json:"isInverse"`
}

type AdminGuideDetailDTO struct {
	ID           uuid.UUID             `json:"id"`
	SectorIDs    []uuid.UUID           `json:"sectorIds"`
	TagIDs       []uuid.UUID           `json:"tagIds"`
	Slug         string                `json:"slug"`
	Icon         *string               `json:"icon,omitempty"`
	ImageURL     *string               `json:"imageUrl,omitempty"`
	SortOrder    int                   `json:"sortOrder"`
	Translations []AdminTranslationDTO `json:"translations"`
	Conditions   []AdminConditionDTO   `json:"conditions"`
}

type AdminStepTranslationDTO struct {
	Language          string      `json:"language"`
	Title             string      `json:"title"`
	Description       *string     `json:"description,omitempty"`
	DetailedContent   interface{} `json:"detailedContent,omitempty"`
	RequiredDocuments interface{} `json:"requiredDocuments,omitempty"`
}

type AdminGuideStepDTO struct {
	ID              uuid.UUID                 `json:"id"`
	GuideID         uuid.UUID                 `json:"guideId"`
	Slug            string                    `json:"slug"`
	StepType        entity.StepType           `json:"stepType"`
	SortOrder       int                       `json:"sortOrder"`
	IsOptional      bool                      `json:"isOptional"`
	EstimatedTime   *int                      `json:"estimatedTime,omitempty"`
	DifficultyLevel *int                      `json:"difficultyLevel,omitempty"`
	FeeEstimate     *int                      `json:"feeEstimate,omitempty"`
	EffectiveDate   time.Time                 `json:"effectiveDate"`
	ExpiryDate      *time.Time                `json:"expiryDate,omitempty"`
	Translations    []AdminStepTranslationDTO `json:"translations"`
}

// --- Mappers ---

func ToCreateGuideInput(body CreateGuideRequest) usecase.CreateGuideInput {
	translations := make([]usecase.TranslationInput, 0, len(body.Translations))
	for _, t := range body.Translations {
		translations = append(translations, usecase.TranslationInput{
			Language:    t.Language,
			Name:        t.Name,
			Description: t.Description,
		})
	}
	conditions := make([]usecase.ConditionInput, 0, len(body.Conditions))
	for _, c := range body.Conditions {
		conditions = append(conditions, usecase.ConditionInput{
			ConditionType:  c.ConditionType,
			Operator:       c.Operator,
			ConditionValue: c.ConditionValue,
			IsInverse:      c.IsInverse,
		})
	}
	return usecase.CreateGuideInput{
		SectorIDs:    body.SectorIDs,
		TagIDs:       body.TagIDs,
		Slug:         body.Slug,
		Icon:         body.Icon,
		ImageURL:     body.ImageURL,
		SortOrder:    body.SortOrder,
		Translations: translations,
		Conditions:   conditions,
	}
}

func ToUpdateGuideInput(body UpdateGuideRequest) usecase.UpdateGuideInput {
	translations := make([]usecase.TranslationInput, 0, len(body.Translations))
	for _, t := range body.Translations {
		translations = append(translations, usecase.TranslationInput{
			Language:    t.Language,
			Name:        t.Name,
			Description: t.Description,
		})
	}
	conditions := make([]usecase.ConditionInput, 0, len(body.Conditions))
	for _, c := range body.Conditions {
		conditions = append(conditions, usecase.ConditionInput{
			ConditionType:  c.ConditionType,
			Operator:       c.Operator,
			ConditionValue: c.ConditionValue,
			IsInverse:      c.IsInverse,
		})
	}
	merge := body.TranslationMode != nil && *body.TranslationMode == "merge"
	return usecase.UpdateGuideInput{
		SectorIDs:         body.SectorIDs,
		TagIDs:            body.TagIDs,
		Slug:              body.Slug,
		Icon:              body.Icon,
		ImageURL:          body.ImageURL,
		SortOrder:         body.SortOrder,
		Translations:      translations,
		TranslationsMerge: merge,
		Conditions:        conditions,
	}
}

func ToCreateStepInput(body CreateStepRequest) usecase.CreateStepInput {
	translations := make([]usecase.StepTranslationInput, 0, len(body.Translations))
	for _, t := range body.Translations {
		translations = append(translations, usecase.StepTranslationInput{
			Language:          t.Language,
			Title:             t.Title,
			Description:       t.Description,
			DetailedContent:   t.DetailedContent,
			RequiredDocuments: t.RequiredDocuments,
		})
	}
	conditions := make([]usecase.ConditionInput, 0, len(body.Conditions))
	for _, c := range body.Conditions {
		conditions = append(conditions, usecase.ConditionInput{
			ConditionType:  c.ConditionType,
			Operator:       c.Operator,
			ConditionValue: c.ConditionValue,
			IsInverse:      c.IsInverse,
		})
	}
	dependencies := make([]usecase.DependencyInput, 0, len(body.Dependencies))
	for _, d := range body.Dependencies {
		dependencies = append(dependencies, usecase.DependencyInput{
			RequiredStepID: d.RequiredStepID,
			DependencyType: d.DependencyType,
		})
	}
	return usecase.CreateStepInput{
		GuideID:         body.GuideID,
		Slug:            body.Slug,
		StepType:        body.StepType,
		SortOrder:       body.SortOrder,
		IsOptional:      body.IsOptional,
		EstimatedTime:   body.EstimatedTime,
		DifficultyLevel: body.DifficultyLevel,
		FeeEstimate:     body.FeeEstimate,
		EffectiveDate:   body.EffectiveDate,
		ExpiryDate:      body.ExpiryDate,
		Translations:    translations,
		Conditions:      conditions,
		Dependencies:    dependencies,
	}
}

func ToUpdateStepInput(body UpdateStepRequest) usecase.UpdateStepInput {
	translations := make([]usecase.StepTranslationInput, 0, len(body.Translations))
	for _, t := range body.Translations {
		translations = append(translations, usecase.StepTranslationInput{
			Language:          t.Language,
			Title:             t.Title,
			Description:       t.Description,
			DetailedContent:   t.DetailedContent,
			RequiredDocuments: t.RequiredDocuments,
		})
	}
	conditions := make([]usecase.ConditionInput, 0, len(body.Conditions))
	for _, c := range body.Conditions {
		conditions = append(conditions, usecase.ConditionInput{
			ConditionType:  c.ConditionType,
			Operator:       c.Operator,
			ConditionValue: c.ConditionValue,
			IsInverse:      c.IsInverse,
		})
	}
	dependencies := make([]usecase.DependencyInput, 0, len(body.Dependencies))
	for _, d := range body.Dependencies {
		dependencies = append(dependencies, usecase.DependencyInput{
			RequiredStepID: d.RequiredStepID,
			DependencyType: d.DependencyType,
		})
	}
	merge := body.TranslationMode != nil && *body.TranslationMode == "merge"
	return usecase.UpdateStepInput{
		Slug:              body.Slug,
		StepType:          body.StepType,
		SortOrder:         body.SortOrder,
		IsOptional:        body.IsOptional,
		EstimatedTime:     body.EstimatedTime,
		DifficultyLevel:   body.DifficultyLevel,
		FeeEstimate:       body.FeeEstimate,
		EffectiveDate:     body.EffectiveDate,
		ExpiryDate:        body.ExpiryDate,
		Translations:      translations,
		TranslationsMerge: merge,
		Conditions:        conditions,
		Dependencies:      dependencies,
	}
}

func ToStepVersionDTO(v *entity.GuideStepVersion) StepVersionDTO {
	return StepVersionDTO{
		ID:            v.ID,
		Version:       v.Version,
		StepID:        v.StepID,
		EffectiveDate: v.EffectiveDate,
		CreatedAt:     v.CreatedAt,
	}
}

func ToAdminQueryOptions(q AdminPaginationQuery) query.QueryOptions {
	opts := query.QueryOptions{}
	if q.Page > 0 {
		opts.Page = q.Page
	}
	if q.PageSize > 0 {
		opts.PageSize = q.PageSize
	}
	opts.Search = q.Search
	opts.SortBy = q.SortBy
	opts.SortOrder = q.SortOrder
	return opts
}

func ToAdminGuideCardDTO(guide *entity.Guide) AdminGuideCardDTO {
	name := ""
	var description *string
	if len(guide.Translations) > 0 {
		name = guide.Translations[0].Name
		description = guide.Translations[0].Description
	}
	return AdminGuideCardDTO{
		ID:          guide.ID,
		Slug:        guide.Slug,
		Name:        name,
		Description: description,
		Icon:        guide.Icon,
		ImageURL:    guide.ImageURL,
		SectorIDs:   guide.SectorIDs,
		TagIDs:      guide.TagIDs,
		SortOrder:   guide.SortOrder,
	}
}

func ToAdminGuideDetailDTO(guide *entity.Guide) AdminGuideDetailDTO {
	translations := make([]AdminTranslationDTO, 0, len(guide.Translations))
	for _, tr := range guide.Translations {
		translations = append(translations, AdminTranslationDTO{
			Language:    tr.Language,
			Name:        tr.Name,
			Description: tr.Description,
		})
	}

	conditions := make([]AdminConditionDTO, 0, len(guide.Conditions))
	for _, c := range guide.Conditions {
		conditions = append(conditions, AdminConditionDTO{
			ID:             c.ID,
			ConditionType:  c.ConditionType,
			Operator:       c.Operator,
			ConditionValue: c.ConditionValue,
			IsInverse:      c.IsInverse,
		})
	}

	return AdminGuideDetailDTO{
		ID:           guide.ID,
		SectorIDs:    guide.SectorIDs,
		TagIDs:       guide.TagIDs,
		Slug:         guide.Slug,
		Icon:         guide.Icon,
		ImageURL:     guide.ImageURL,
		SortOrder:    guide.SortOrder,
		Translations: translations,
		Conditions:   conditions,
	}
}

func ToAdminGuideStepDTO(step *entity.GuideStep) AdminGuideStepDTO {
	translations := make([]AdminStepTranslationDTO, 0, len(step.Translations))
	for _, tr := range step.Translations {
		translations = append(translations, AdminStepTranslationDTO{
			Language:          tr.Language,
			Title:             tr.Title,
			Description:       tr.Description,
			DetailedContent:   tr.DetailedContent,
			RequiredDocuments: tr.RequiredDocuments,
		})
	}

	return AdminGuideStepDTO{
		ID:              step.ID,
		GuideID:         step.GuideID,
		Slug:            step.Slug,
		StepType:        step.StepType,
		SortOrder:       step.SortOrder,
		IsOptional:      step.IsOptional,
		EstimatedTime:   step.EstimatedTime,
		DifficultyLevel: step.DifficultyLevel,
		FeeEstimate:     step.FeeEstimate,
		EffectiveDate:   step.EffectiveDate,
		ExpiryDate:      step.ExpiryDate,
		Translations:    translations,
	}
}
