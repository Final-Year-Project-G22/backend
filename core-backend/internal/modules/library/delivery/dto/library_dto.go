package dto

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
	"github.com/google/uuid"
)

// --- List Categories ---

type LibraryListCategoriesInput struct {
	Locale string `query:"locale,omitempty" doc:"Language code for localized names"`
}

type CategoryNodeResponse struct {
	ID        uuid.UUID              `json:"id" doc:"Category ID"`
	Name      string                 `json:"name" doc:"Category name"`
	Slug      string                 `json:"slug" doc:"Category slug"`
	Icon      *string                `json:"icon,omitempty" doc:"Icon name"`
	SortOrder int                    `json:"sortOrder" doc:"Display order"`
	Children  []CategoryNodeResponse `json:"children,omitempty" doc:"Child categories"`
}

type LibraryListCategoriesOutput struct {
	Body []CategoryNodeResponse
}

// --- List Template Groups ---

type ListTemplateGroupsInput struct {
	CategoryID string `query:"categoryId,omitempty" doc:"Filter by category ID"`
	Format     string `query:"format,omitempty" doc:"Filter by format (pdf, docx, xlsx, interactive_form)"`
	Search     string `query:"search,omitempty" doc:"Search by name or description"`
	Page       int    `query:"page" doc:"Page number" default:"1"`
	PageSize   int    `query:"pageSize" doc:"Items per page" default:"20" maximum:"100"`
}

type TemplateGroupCardResponse struct {
	ID            uuid.UUID             `json:"id" doc:"Group ID"`
	Name          string                `json:"name" doc:"Group name"`
	Slug          string                `json:"slug" doc:"Group slug"`
	CategoryID    uuid.UUID             `json:"categoryId" doc:"Category ID"`
	Format        entity.TemplateFormat `json:"format" doc:"Template format"`
	TierAccess    entity.TierAccess     `json:"tierAccess" doc:"Tier access level"`
	SortOrder     int                   `json:"sortOrder" doc:"Display order"`
	ThumbnailURL  *string               `json:"thumbnailUrl,omitempty" doc:"Thumbnail URL"`
	DownloadCount int                   `json:"downloadCount" doc:"Total downloads"`
	Languages     []string              `json:"languages" doc:"Available language codes"`
}

type ListTemplateGroupsResponseBody struct {
	Data       []TemplateGroupCardResponse `json:"data" doc:"Template groups"`
	Total      int64                       `json:"total" doc:"Total count"`
	Page       int                         `json:"page" doc:"Current page"`
	PageSize   int                         `json:"pageSize" doc:"Items per page"`
	TotalPages int                         `json:"totalPages" doc:"Total pages"`
}

type ListTemplateGroupsOutput struct {
	Body ListTemplateGroupsResponseBody
}

// --- Get Template Group ---

type GetTemplateGroupBySlugInput struct {
	Slug   string `path:"slug" doc:"Template group slug"`
	Locale string `query:"locale,omitempty" doc:"Language code"`
}

type LanguageVariantResponse struct {
	Language    string  `json:"language" doc:"Language code"`
	Title       string  `json:"title" doc:"Template title"`
	Description *string `json:"description,omitempty" doc:"Template description"`
	FileSize    int64   `json:"fileSize" doc:"File size in bytes"`
	ContentType string  `json:"contentType" doc:"MIME type"`
	Version     int     `json:"version" doc:"Version number"`
}

type UserTemplateGroupDetailResponse struct {
	ID              uuid.UUID                 `json:"id" doc:"Group ID"`
	Name            string                    `json:"name" doc:"Group name"`
	Slug            string                    `json:"slug" doc:"Group slug"`
	CategoryID      uuid.UUID                 `json:"categoryId" doc:"Category ID"`
	Format          entity.TemplateFormat     `json:"format" doc:"Template format"`
	TierAccess      entity.TierAccess         `json:"tierAccess" doc:"Tier access level"`
	RequiresAuth    bool                      `json:"requiresAuth" doc:"Requires authentication"`
	SortOrder       int                       `json:"sortOrder" doc:"Display order"`
	DefaultLanguage string                    `json:"defaultLanguage" doc:"Default language code"`
	ThumbnailURL    *string                   `json:"thumbnailUrl,omitempty" doc:"Thumbnail URL"`
	DownloadCount   int                       `json:"downloadCount" doc:"Total downloads"`
	Languages       []LanguageVariantResponse `json:"languages" doc:"Available language variants"`
}

type GetTemplateGroupDetailOutput struct {
	Body UserTemplateGroupDetailResponse
}

// --- Download Template ---

type DownloadTemplateInput struct {
	Slug     string `path:"slug" doc:"Template group slug"`
	Language string `query:"language,omitempty" doc:"Language code"`
}

type DownloadTemplateOutput struct {
	Body DownloadTemplateResponseBody
}

type DownloadTemplateResponseBody struct {
	PresignedURL string `json:"presignedUrl" doc:"Temporary download URL"`
	ExpiresAt    string `json:"expiresAt" doc:"Expiry time"`
	Filename     string `json:"filename" doc:"Suggested filename"`
}

// --- List My Downloads ---

type ListMyDownloadsInput struct {
	Page     int `query:"page" doc:"Page number" default:"1"`
	PageSize int `query:"pageSize" doc:"Items per page" default:"20" maximum:"100"`
}

type MyDownloadResponse struct {
	ID           uuid.UUID `json:"id" doc:"Download log ID"`
	TemplateID   uuid.UUID `json:"templateId" doc:"Template ID"`
	GroupID      uuid.UUID `json:"groupId" doc:"Group ID"`
	DownloadedAt string    `json:"downloadedAt" doc:"Download timestamp"`
}

type ListMyDownloadsResponseBody struct {
	Data       []MyDownloadResponse `json:"data" doc:"Downloads"`
	Total      int64                `json:"total" doc:"Total count"`
	Page       int                  `json:"page" doc:"Current page"`
	PageSize   int                  `json:"pageSize" doc:"Items per page"`
	TotalPages int                  `json:"totalPages" doc:"Total pages"`
}

type ListMyDownloadsOutput struct {
	Body ListMyDownloadsResponseBody
}

// --- Mappers ---

func ToCategoryNodeResponse(cat *entity.LibraryCategory, children []CategoryNodeResponse) CategoryNodeResponse {
	return CategoryNodeResponse{
		ID:        cat.ID,
		Name:      cat.Name,
		Slug:      cat.Slug,
		Icon:      cat.Icon,
		SortOrder: cat.SortOrder,
		Children:  children,
	}
}

func ToTemplateGroupCardResponse(group *entity.LibraryTemplateGroup) TemplateGroupCardResponse {
	langs := make([]string, 0)
	for _, t := range group.Templates {
		if t.IsActive {
			langs = append(langs, t.Language)
		}
	}
	return TemplateGroupCardResponse{
		ID:            group.ID,
		Name:          group.Name,
		Slug:          group.Slug,
		CategoryID:    group.CategoryID,
		Format:        group.Format,
		TierAccess:    group.TierAccess,
		SortOrder:     group.SortOrder,
		ThumbnailURL:  group.ThumbnailURL,
		DownloadCount: group.DownloadCount,
		Languages:     langs,
	}
}

func ToGroupDetailResponse(group *entity.LibraryTemplateGroup, templates []*entity.LibraryTemplate) UserTemplateGroupDetailResponse {
	langs := make([]LanguageVariantResponse, 0)
	for _, t := range templates {
		langs = append(langs, LanguageVariantResponse{
			Language:    t.Language,
			Title:       t.Title,
			Description: t.Description,
			FileSize:    t.FileSize,
			ContentType: t.ContentType,
			Version:     t.Version,
		})
	}
	return UserTemplateGroupDetailResponse{
		ID:              group.ID,
		Name:            group.Name,
		Slug:            group.Slug,
		CategoryID:      group.CategoryID,
		Format:          group.Format,
		TierAccess:      group.TierAccess,
		RequiresAuth:    group.RequiresAuth,
		SortOrder:       group.SortOrder,
		DefaultLanguage: group.DefaultLanguage,
		ThumbnailURL:    group.ThumbnailURL,
		DownloadCount:   group.DownloadCount,
		Languages:       langs,
	}
}

func ToMyDownloadResponse(dl *entity.LibraryTemplateDownload) MyDownloadResponse {
	downloadedAt := ""
	if dl.CreatedAt != nil {
		downloadedAt = dl.CreatedAt.Format("2006-01-02T15:04:05Z")
	}
	return MyDownloadResponse{
		ID:           dl.ID,
		TemplateID:   dl.TemplateID,
		GroupID:      dl.GroupID,
		DownloadedAt: downloadedAt,
	}
}
