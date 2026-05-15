package usecase

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/library/domain/entity"
	"github.com/google/uuid"
)

type CreateCategoryInput struct {
	Name             string
	Slug             string
	Icon             *string
	SortOrder        int
	ParentCategoryID *uuid.UUID
}

type UpdateCategoryInput struct {
	Name      *string
	Slug      *string
	Icon      *string
	SortOrder *int
	IsActive  *bool
}

type CreateCategoryTranslationInput struct {
	CategoryID  uuid.UUID
	Language    string
	Name        string
	Description *string
}

type UpdateCategoryTranslationInput struct {
	Name        *string
	Description *string
}

type CreateTemplateGroupInput struct {
	Name              string
	Description       *string
	Slug              string
	CategoryID        uuid.UUID
	Format            entity.TemplateFormat
	TierAccess        entity.TierAccess
	RequiresAuth      bool
	SortOrder         int
	DefaultLanguage   string
	ThumbnailBytes    []byte
	ThumbnailFilename *string
	ThumbnailURL      *string
}

type UpdateTemplateGroupInput struct {
	Name              *string
	Description       *string
	Slug              *string
	CategoryID        *uuid.UUID
	Format            *entity.TemplateFormat
	TierAccess        *entity.TierAccess
	RequiresAuth      *bool
	SortOrder         *int
	DefaultLanguage   *string
	IsActive          *bool
	ThumbnailBytes    []byte
	ThumbnailFilename *string
	ThumbnailURL      *string
}

type CreateTemplateInput struct {
	GroupID     uuid.UUID
	Language    string
	Title       string
	Description *string
	FileBytes   []byte
	Filename    string
	FileKey     string
	FileURL     *string
	FileSize    int64
	ContentType string
}

type UpdateTemplateInput struct {
	Title       *string
	Description *string
	FileBytes   []byte
	Filename    *string
	IsActive    *bool
	FileKey     *string
	FileURL     *string
	FileSize    *int64
	ContentType *string
}

type CreateInteractiveFormInput struct {
	TemplateID  uuid.UUID
	Name        string
	Description *string
	FormLayout  map[string]interface{}
}

type UpdateInteractiveFormInput struct {
	Name        *string
	Description *string
	FormLayout  *map[string]interface{}
}

type DownloadInput struct {
	GroupID   uuid.UUID
	Language  *string
	AccountID *uuid.UUID
}

type DownloadOutput struct {
	PresignedURL string
	ExpiresAt    string
	Filename     string
	ContentType  string
}

type PreviewInput struct {
	GroupID   uuid.UUID
	Language  *string
	AccountID *uuid.UUID
}

type PreviewOutput struct {
	PresignedURL string
	ExpiresAt    string
	Filename     string
	ContentType  string
}

type CreateTemplateUploadIntentInput struct {
	GroupID     uuid.UUID
	Language    string
	Title       string
	Description *string
	FileName    string
	ContentType string
	FileSize    int64
}

type CreateTemplateUploadIntentOutput struct {
	UploadURL string
	Method    string
	Headers   map[string]string
	FileKey   string
	ExpiresAt string
}
