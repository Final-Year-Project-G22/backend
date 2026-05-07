package dto

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// --- Category ---

type CreateLibraryCategoryRequest struct {
	Name             string     `json:"name" doc:"Category name" minLength:"1" maxLength:"200"`
	Slug             string     `json:"slug" doc:"URL-friendly identifier" minLength:"1" maxLength:"200"`
	Icon             *string    `json:"icon,omitempty" doc:"Icon name"`
	SortOrder        int        `json:"sortOrder" doc:"Display order"`
	ParentCategoryID *uuid.UUID `json:"parentCategoryId,omitempty" doc:"Parent category ID"`
}

type CreateLibraryCategoryInput struct {
	Body CreateLibraryCategoryRequest
}

type CreateLibraryCategoryOutput struct {
	Body struct {
		ID uuid.UUID `json:"id" doc:"Created category ID"`
	}
}

type LibraryGetCategoryInput struct {
	ID uuid.UUID `path:"id" doc:"Category ID"`
}

type CategoryDetailResponse struct {
	ID               uuid.UUID                 `json:"id" doc:"Category ID"`
	Name             string                    `json:"name" doc:"Category name"`
	Slug             string                    `json:"slug" doc:"Category slug"`
	Icon             *string                   `json:"icon,omitempty" doc:"Icon name"`
	SortOrder        int                       `json:"sortOrder" doc:"Display order"`
	ParentCategoryID *uuid.UUID                `json:"parentCategoryId,omitempty" doc:"Parent category ID"`
	IsActive         bool                      `json:"isActive" doc:"Active flag"`
	Translations     []CategoryTranslationItem `json:"translations,omitempty" doc:"Category translations"`
}

type CategoryTranslationItem struct {
	Language    string  `json:"language" doc:"Language code"`
	Name        string  `json:"name" doc:"Translated name"`
	Description *string `json:"description,omitempty" doc:"Translated description"`
}

type LibraryGetCategoryOutput struct {
	Body CategoryDetailResponse
}

type UpdateLibraryCategoryRequest struct {
	Name      *string `json:"name,omitempty" doc:"Category name"`
	Slug      *string `json:"slug,omitempty" doc:"URL-friendly identifier"`
	Icon      *string `json:"icon,omitempty" doc:"Icon name"`
	SortOrder *int    `json:"sortOrder,omitempty" doc:"Display order"`
	IsActive  *bool   `json:"isActive,omitempty" doc:"Active flag"`
}

type UpdateLibraryCategoryInput struct {
	ID   uuid.UUID `path:"id" doc:"Category ID"`
	Body UpdateLibraryCategoryRequest
}

type UpdateLibraryCategoryOutput struct {
	Body CategoryDetailResponse
}

type LibraryDeleteCategoryInput struct {
	ID uuid.UUID `path:"id" doc:"Category ID"`
}

type LibraryDeleteCategoryOutput struct {
	Body struct {
		Message string `json:"message" doc:"Success message"`
	}
}

type ListAllCategoriesInput struct {
	IncludeInactive bool `query:"includeInactive" doc:"Include inactive categories" default:"false"`
}

type CategorySummaryResponse struct {
	ID        uuid.UUID `json:"id" doc:"Category ID"`
	Name      string    `json:"name" doc:"Category name"`
	Slug      string    `json:"slug" doc:"Category slug"`
	Icon      *string   `json:"icon,omitempty" doc:"Icon name"`
	SortOrder int       `json:"sortOrder" doc:"Display order"`
	IsActive  bool      `json:"isActive" doc:"Active flag"`
}

type ListAllCategoriesOutput struct {
	Body []CategorySummaryResponse
}

// --- Category Translations ---

type AddCategoryTranslationRequest struct {
	Language    string  `json:"language" doc:"Language code" minLength:"2" maxLength:"10"`
	Name        string  `json:"name" doc:"Translated name" minLength:"1" maxLength:"200"`
	Description *string `json:"description,omitempty" doc:"Translated description"`
}

type AddCategoryTranslationInput struct {
	ID   uuid.UUID `path:"id" doc:"Category ID"`
	Body AddCategoryTranslationRequest
}

type AddCategoryTranslationOutput struct {
	Body struct {
		ID uuid.UUID `json:"id" doc:"Translation ID"`
	}
}

type UpdateCategoryTranslationRequest struct {
	Name        *string `json:"name,omitempty" doc:"Translated name"`
	Description *string `json:"description,omitempty" doc:"Translated description"`
}

type UpdateCategoryTranslationInput struct {
	CategoryID uuid.UUID `path:"id" doc:"Category ID"`
	Language   string    `path:"lang" doc:"Language code"`
	Body       UpdateCategoryTranslationRequest
}

type UpdateCategoryTranslationOutput struct {
	Body struct {
		Message string `json:"message" doc:"Success message"`
	}
}

type DeleteCategoryTranslationInput struct {
	CategoryID uuid.UUID `path:"id" doc:"Category ID"`
	Language   string    `path:"lang" doc:"Language code"`
}

type DeleteCategoryTranslationOutput struct {
	Body struct {
		Message string `json:"message" doc:"Success message"`
	}
}

// --- Template Groups ---

type CreateTemplateGroupRequest struct {
	Name            string                `json:"name" doc:"Group name" minLength:"1" maxLength:"200"`
	Description     *string               `json:"description,omitempty" doc:"Group description"`
	Slug            string                `json:"slug" doc:"URL-friendly identifier" minLength:"1" maxLength:"200"`
	CategoryID      uuid.UUID             `json:"categoryId" doc:"Category ID"`
	Format          entity.TemplateFormat `json:"format" doc:"Template format"`
	TierAccess      entity.TierAccess     `json:"tierAccess" doc:"Tier access level"`
	RequiresAuth    bool                  `json:"requiresAuth" doc:"Requires authentication"`
	SortOrder       int                   `json:"sortOrder" doc:"Display order"`
	DefaultLanguage string                `json:"defaultLanguage" doc:"Default language code"`
}

type CreateTemplateGroupInput struct {
	Body CreateTemplateGroupRequest
}

type CreateTemplateGroupOutput struct {
	Body struct {
		ID uuid.UUID `json:"id" doc:"Created group ID"`
	}
}

type GetTemplateGroupInput struct {
	GroupID uuid.UUID `path:"groupId" doc:"Group ID"`
}

type TemplateGroupDetailResponse struct {
	ID              uuid.UUID             `json:"id" doc:"Group ID"`
	Name            string                `json:"name" doc:"Group name"`
	Description     *string               `json:"description,omitempty" doc:"Group description"`
	Slug            string                `json:"slug" doc:"Group slug"`
	CategoryID      uuid.UUID             `json:"categoryId" doc:"Category ID"`
	Format          entity.TemplateFormat `json:"format" doc:"Template format"`
	TierAccess      entity.TierAccess     `json:"tierAccess" doc:"Tier access level"`
	RequiresAuth    bool                  `json:"requiresAuth" doc:"Requires authentication"`
	IsActive        bool                  `json:"isActive" doc:"Active flag"`
	SortOrder       int                   `json:"sortOrder" doc:"Display order"`
	DefaultLanguage string                `json:"defaultLanguage" doc:"Default language code"`
	ThumbnailURL    *string               `json:"thumbnailUrl,omitempty" doc:"Thumbnail URL"`
	DownloadCount   int                   `json:"downloadCount" doc:"Total downloads"`
	CreatedBy       uuid.UUID             `json:"createdBy" doc:"Creator account ID"`
	CreatedAt       time.Time             `json:"createdAt" doc:"Creation time"`
	Templates       []TemplateItem        `json:"templates,omitempty" doc:"Language variants"`
}

type TemplateItem struct {
	ID          uuid.UUID `json:"id" doc:"Template ID"`
	Language    string    `json:"language" doc:"Language code"`
	Title       string    `json:"title" doc:"Template title"`
	Description *string   `json:"description,omitempty" doc:"Template description"`
	FileSize    int64     `json:"fileSize" doc:"File size in bytes"`
	ContentType string    `json:"contentType" doc:"MIME type"`
	Version     int       `json:"version" doc:"Version number"`
	IsActive    bool      `json:"isActive" doc:"Active flag"`
}

type GetTemplateGroupOutput struct {
	Body TemplateGroupDetailResponse
}

type UpdateTemplateGroupRequest struct {
	Name            *string                `json:"name,omitempty" doc:"Group name"`
	Description     *string                `json:"description,omitempty" doc:"Group description"`
	Slug            *string                `json:"slug,omitempty" doc:"Group slug"`
	CategoryID      *uuid.UUID             `json:"categoryId,omitempty" doc:"Category ID"`
	Format          *entity.TemplateFormat `json:"format,omitempty" doc:"Template format"`
	TierAccess      *entity.TierAccess     `json:"tierAccess,omitempty" doc:"Tier access level"`
	RequiresAuth    *bool                  `json:"requiresAuth,omitempty" doc:"Requires authentication"`
	SortOrder       *int                   `json:"sortOrder,omitempty" doc:"Display order"`
	DefaultLanguage *string                `json:"defaultLanguage,omitempty" doc:"Default language code"`
	IsActive        *bool                  `json:"isActive,omitempty" doc:"Active flag"`
}

type UpdateTemplateGroupInput struct {
	GroupID uuid.UUID `path:"groupId" doc:"Group ID"`
	Body    UpdateTemplateGroupRequest
}

type UpdateTemplateGroupOutput struct {
	Body TemplateGroupDetailResponse
}

type DeleteTemplateGroupInput struct {
	GroupID uuid.UUID `path:"groupId" doc:"Group ID"`
}

type DeleteTemplateGroupOutput struct {
	Body struct {
		Message string `json:"message" doc:"Success message"`
	}
}

type ListAllTemplateGroupsInput struct {
	CategoryID string `query:"categoryId,omitempty" doc:"Filter by category ID"`
	Page       int    `query:"page" doc:"Page number" default:"1"`
	PageSize   int    `query:"pageSize" doc:"Items per page" default:"20" maximum:"100"`
}

type TemplateGroupSummaryResponse struct {
	ID            uuid.UUID             `json:"id" doc:"Group ID"`
	Name          string                `json:"name" doc:"Group name"`
	Slug          string                `json:"slug" doc:"Group slug"`
	CategoryID    uuid.UUID             `json:"categoryId" doc:"Category ID"`
	Format        entity.TemplateFormat `json:"format" doc:"Template format"`
	TierAccess    entity.TierAccess     `json:"tierAccess" doc:"Tier access level"`
	IsActive      bool                  `json:"isActive" doc:"Active flag"`
	SortOrder     int                   `json:"sortOrder" doc:"Display order"`
	ThumbnailURL  *string               `json:"thumbnailUrl,omitempty" doc:"Thumbnail URL"`
	DownloadCount int                   `json:"downloadCount" doc:"Total downloads"`
}

type ListAllTemplateGroupsOutput struct {
	Body []TemplateGroupSummaryResponse
}

// --- Templates ---

type CreateTemplateFormData struct {
	File        huma.FormFile `form:"file" required:"true" doc:"Template file"`
	Language    string        `form:"language" required:"true" doc:"Language code"`
	Title       string        `form:"title" required:"true" doc:"Template title"`
	Description string        `form:"description" doc:"Template description"`
}

type LibraryCreateTemplateInput struct {
	GroupID uuid.UUID `path:"groupId" doc:"Group ID"`
	RawBody huma.MultipartFormFiles[CreateTemplateFormData]
}

type LibraryCreateTemplateOutput struct {
	Body struct {
		ID uuid.UUID `json:"id" doc:"Created template ID"`
	}
}

type LibraryGetTemplateInput struct {
	TemplateID uuid.UUID `path:"templateId" doc:"Template ID"`
}

type LibraryTemplateDetailResponse struct {
	ID          uuid.UUID `json:"id" doc:"Template ID"`
	GroupID     uuid.UUID `json:"groupId" doc:"Group ID"`
	Language    string    `json:"language" doc:"Language code"`
	Title       string    `json:"title" doc:"Template title"`
	Description *string   `json:"description,omitempty" doc:"Template description"`
	FileURL     *string   `json:"fileUrl,omitempty" doc:"File URL"`
	FileSize    int64     `json:"fileSize" doc:"File size in bytes"`
	ContentType string    `json:"contentType" doc:"MIME type"`
	Version     int       `json:"version" doc:"Version number"`
	IsActive    bool      `json:"isActive" doc:"Active flag"`
}

type LibraryGetTemplateOutput struct {
	Body LibraryTemplateDetailResponse
}

type UpdateTemplateFormData struct {
	File        huma.FormFile `form:"file" doc:"Template file (optional)"`
	Title       string        `form:"title" doc:"Template title"`
	Description string        `form:"description" doc:"Template description"`
	IsActive    string        `form:"isActive" doc:"Active flag (true/false)"`
}

type LibraryUpdateTemplateInput struct {
	TemplateID uuid.UUID `path:"templateId" doc:"Template ID"`
	RawBody    huma.MultipartFormFiles[UpdateTemplateFormData]
}

type LibraryUpdateTemplateOutput struct {
	Body LibraryTemplateDetailResponse
}

type LibraryDeleteTemplateInput struct {
	TemplateID uuid.UUID `path:"templateId" doc:"Template ID"`
}

type LibraryDeleteTemplateOutput struct {
	Body struct {
		Message string `json:"message" doc:"Success message"`
	}
}

type ListTemplatesByGroupInput struct {
	GroupID uuid.UUID `path:"groupId" doc:"Group ID"`
}

type ListTemplatesByGroupOutput struct {
	Body []TemplateItem
}

// --- Interactive Forms ---

type CreateInteractiveFormRequest struct {
	Name        string                 `json:"name" doc:"Form name" minLength:"1" maxLength:"100"`
	Description *string                `json:"description,omitempty" doc:"Form description"`
	FormLayout  map[string]interface{} `json:"formLayout" doc:"Form layout JSON"`
}

type CreateInteractiveFormInput struct {
	TemplateID uuid.UUID `path:"templateId" doc:"Template ID"`
	Body       CreateInteractiveFormRequest
}

type CreateInteractiveFormOutput struct {
	Body struct {
		ID uuid.UUID `json:"id" doc:"Created form ID"`
	}
}

type GetInteractiveFormInput struct {
	TemplateID uuid.UUID `path:"templateId" doc:"Template ID"`
}

type InteractiveFormDetailResponse struct {
	ID          uuid.UUID              `json:"id" doc:"Form ID"`
	TemplateID  uuid.UUID              `json:"templateId" doc:"Template ID"`
	Name        string                 `json:"name" doc:"Form name"`
	Description *string                `json:"description,omitempty" doc:"Form description"`
	FormLayout  map[string]interface{} `json:"formLayout" doc:"Form layout JSON"`
	Version     int                    `json:"version" doc:"Version number"`
	IsActive    bool                   `json:"isActive" doc:"Active flag"`
}

type GetInteractiveFormOutput struct {
	Body InteractiveFormDetailResponse
}

type UpdateInteractiveFormRequest struct {
	Name        *string                 `json:"name,omitempty" doc:"Form name"`
	Description *string                 `json:"description,omitempty" doc:"Form description"`
	FormLayout  *map[string]interface{} `json:"formLayout,omitempty" doc:"Form layout JSON"`
}

type UpdateInteractiveFormInput struct {
	ID   uuid.UUID `path:"id" doc:"Form ID"`
	Body UpdateInteractiveFormRequest
}

type UpdateInteractiveFormOutput struct {
	Body InteractiveFormDetailResponse
}

type DeleteInteractiveFormInput struct {
	ID uuid.UUID `path:"id" doc:"Form ID"`
}

type DeleteInteractiveFormOutput struct {
	Body struct {
		Message string `json:"message" doc:"Success message"`
	}
}

// --- Download Logs ---

type ListDownloadLogsInput struct {
	GroupID  string `query:"groupId,omitempty" doc:"Filter by template group ID"`
	Page     int    `query:"page" doc:"Page number" default:"1"`
	PageSize int    `query:"pageSize" doc:"Items per page" default:"20" maximum:"100"`
}

type DownloadLogResponse struct {
	ID           uuid.UUID `json:"id" doc:"Download log ID"`
	AccountID    uuid.UUID `json:"accountId" doc:"Account ID"`
	TemplateID   uuid.UUID `json:"templateId" doc:"Template ID"`
	GroupID      uuid.UUID `json:"groupId" doc:"Group ID"`
	DownloadedAt time.Time `json:"downloadedAt" doc:"Download timestamp"`
}

type DownloadLogListResponse struct {
	Data       []DownloadLogResponse `json:"data" doc:"Download logs"`
	Total      int64                 `json:"total" doc:"Total count"`
	Page       int                   `json:"page" doc:"Current page"`
	PageSize   int                   `json:"pageSize" doc:"Items per page"`
	TotalPages int                   `json:"totalPages" doc:"Total pages"`
}

type ListDownloadLogsOutput struct {
	Body DownloadLogListResponse
}

// --- Thumbnail Upload ---

type UploadThumbnailFormData struct {
	File huma.FormFile `form:"file" doc:"Thumbnail image"`
}

type UploadThumbnailInput struct {
	ID      uuid.UUID `path:"id" doc:"Group ID"`
	RawBody huma.MultipartFormFiles[UploadThumbnailFormData]
}

type UploadThumbnailOutput struct {
	Body struct {
		ThumbnailURL string `json:"thumbnailUrl" doc:"Uploaded thumbnail URL"`
	}
}

// --- Mappers ---

func ToQueryOptions(page, pageSize int) query.QueryOptions {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return query.QueryOptions{
		Page:     page,
		PageSize: pageSize,
	}
}

func ToCategoryDetailResponse(cat *entity.LibraryCategory) CategoryDetailResponse {
	resp := CategoryDetailResponse{
		ID:               cat.ID,
		Name:             cat.Name,
		Slug:             cat.Slug,
		Icon:             cat.Icon,
		SortOrder:        cat.SortOrder,
		ParentCategoryID: cat.ParentCategoryID,
		IsActive:         cat.IsActive,
	}
	for _, t := range cat.Translations {
		resp.Translations = append(resp.Translations, CategoryTranslationItem{
			Language:    t.Language,
			Name:        t.Name,
			Description: t.Description,
		})
	}
	return resp
}

func ToCategorySummaryResponse(cat *entity.LibraryCategory) CategorySummaryResponse {
	return CategorySummaryResponse{
		ID:        cat.ID,
		Name:      cat.Name,
		Slug:      cat.Slug,
		Icon:      cat.Icon,
		SortOrder: cat.SortOrder,
		IsActive:  cat.IsActive,
	}
}

func ToTemplateGroupDetailResponse(group *entity.LibraryTemplateGroup) TemplateGroupDetailResponse {
	resp := TemplateGroupDetailResponse{
		ID:              group.ID,
		Name:            group.Name,
		Description:     group.Description,
		Slug:            group.Slug,
		CategoryID:      group.CategoryID,
		Format:          group.Format,
		TierAccess:      group.TierAccess,
		RequiresAuth:    group.RequiresAuth,
		IsActive:        group.IsActive,
		SortOrder:       group.SortOrder,
		DefaultLanguage: group.DefaultLanguage,
		ThumbnailURL:    group.ThumbnailURL,
		DownloadCount:   group.DownloadCount,
		CreatedBy:       group.CreatedBy,
		CreatedAt:       *group.CreatedAt,
	}
	for _, t := range group.Templates {
		resp.Templates = append(resp.Templates, ToTemplateItem(t))
	}
	return resp
}

func ToTemplateItem(tmpl entity.LibraryTemplate) TemplateItem {
	return TemplateItem{
		ID:          tmpl.ID,
		Language:    tmpl.Language,
		Title:       tmpl.Title,
		Description: tmpl.Description,
		FileSize:    tmpl.FileSize,
		ContentType: tmpl.ContentType,
		Version:     tmpl.Version,
		IsActive:    tmpl.IsActive,
	}
}

func ToTemplateGroupSummaryResponse(group *entity.LibraryTemplateGroup) TemplateGroupSummaryResponse {
	return TemplateGroupSummaryResponse{
		ID:            group.ID,
		Name:          group.Name,
		Slug:          group.Slug,
		CategoryID:    group.CategoryID,
		Format:        group.Format,
		TierAccess:    group.TierAccess,
		IsActive:      group.IsActive,
		SortOrder:     group.SortOrder,
		ThumbnailURL:  group.ThumbnailURL,
		DownloadCount: group.DownloadCount,
	}
}

func ToLibraryTemplateDetailResponse(tmpl *entity.LibraryTemplate) LibraryTemplateDetailResponse {
	return LibraryTemplateDetailResponse{
		ID:          tmpl.ID,
		GroupID:     tmpl.GroupID,
		Language:    tmpl.Language,
		Title:       tmpl.Title,
		Description: tmpl.Description,
		FileURL:     tmpl.FileURL,
		FileSize:    tmpl.FileSize,
		ContentType: tmpl.ContentType,
		Version:     tmpl.Version,
		IsActive:    tmpl.IsActive,
	}
}

func ToInteractiveFormDetailResponse(form *entity.LibraryInteractiveForm) InteractiveFormDetailResponse {
	return InteractiveFormDetailResponse{
		ID:          form.ID,
		TemplateID:  form.TemplateID,
		Name:        form.Name,
		Description: form.Description,
		FormLayout:  form.FormLayout,
		Version:     form.Version,
		IsActive:    form.IsActive,
	}
}

func ToDownloadLogResponse(dl *entity.LibraryTemplateDownload) DownloadLogResponse {
	return DownloadLogResponse{
		ID:           dl.ID,
		AccountID:    dl.AccountID,
		TemplateID:   dl.TemplateID,
		GroupID:      dl.GroupID,
		DownloadedAt: *dl.CreatedAt,
	}
}
