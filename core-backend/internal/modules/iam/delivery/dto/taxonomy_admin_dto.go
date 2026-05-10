package dto

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/google/uuid"
)

// --- Sector ---

type CreateSectorRequest struct {
	Slug      string     `json:"slug" doc:"Sector slug" minLength:"1" maxLength:"100"`
	ParentID  *uuid.UUID `json:"parentId,omitempty" doc:"Parent sector ID"`
	Icon      *string    `json:"icon,omitempty" doc:"Icon identifier" maxLength:"50"`
	SortOrder int        `json:"sortOrder" doc:"Display order"`
	IsActive  bool       `json:"isActive" doc:"Whether sector is active"`
	NameEN    string     `json:"nameEn" doc:"English name" minLength:"1" maxLength:"200"`
	DescEN    *string    `json:"descEn,omitempty" doc:"English description"`
	NameAM    string     `json:"nameAm" doc:"Amharic name" minLength:"1" maxLength:"200"`
	DescAM    *string    `json:"descAm,omitempty" doc:"Amharic description"`
}

type CreateSectorInput struct {
	Body CreateSectorRequest
}

type CreateSectorOutput struct {
	Body CreateSectorResponseBody
}

type CreateSectorResponseBody struct {
	ID uuid.UUID `json:"id" doc:"Created sector ID"`
}

type UpdateSectorRequest struct {
	Slug      *string    `json:"slug,omitempty" doc:"Sector slug" maxLength:"100"`
	ParentID  *uuid.UUID `json:"parentId,omitempty" doc:"Parent sector ID"`
	Icon      *string    `json:"icon,omitempty" doc:"Icon identifier" maxLength:"50"`
	SortOrder *int       `json:"sortOrder,omitempty" doc:"Display order"`
	IsActive  *bool      `json:"isActive,omitempty" doc:"Whether sector is active"`
	NameEN    *string    `json:"nameEn,omitempty" doc:"English name" maxLength:"200"`
	DescEN    *string    `json:"descEn,omitempty" doc:"English description"`
	NameAM    *string    `json:"nameAm,omitempty" doc:"Amharic name" maxLength:"200"`
	DescAM    *string    `json:"descAm,omitempty" doc:"Amharic description"`
}

type UpdateSectorInput struct {
	ID   uuid.UUID `path:"id" doc:"Sector ID"`
	Body UpdateSectorRequest
}

type UpdateSectorOutput struct {
	Body UpdateSectorResponseBody
}

type UpdateSectorResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type GetSectorInput struct {
	ID uuid.UUID `path:"id" doc:"Sector ID"`
}

type GetSectorOutput struct {
	Body SectorResponse
}

type ListSectorsInput struct {
	Page     int    `query:"page" doc:"Page number"`
	PageSize int    `query:"pageSize" doc:"Items per page"`
	Search   string `query:"search" doc:"Search term"`
}

type ListSectorsOutput struct {
	Body ListSectorsResponseBody
}

type ListSectorsResponseBody struct {
	Data       []SectorResponse `json:"data" doc:"Sector list"`
	Total      int64            `json:"total" doc:"Total count"`
	Page       int              `json:"page" doc:"Current page"`
	PageSize   int              `json:"pageSize" doc:"Items per page"`
	TotalPages int              `json:"totalPages" doc:"Total pages"`
}

type SectorResponse struct {
	ID        uuid.UUID  `json:"id" doc:"Sector ID"`
	Slug      string     `json:"slug" doc:"Sector slug"`
	ParentID  *uuid.UUID `json:"parentId,omitempty" doc:"Parent sector ID"`
	Icon      *string    `json:"icon,omitempty" doc:"Icon identifier"`
	SortOrder int        `json:"sortOrder" doc:"Display order"`
	IsActive  bool       `json:"isActive" doc:"Whether sector is active"`
	NameEN    string     `json:"nameEn" doc:"English name"`
	DescEN    *string    `json:"descEn,omitempty" doc:"English description"`
	NameAM    string     `json:"nameAm" doc:"Amharic name"`
	DescAM    *string    `json:"descAm,omitempty" doc:"Amharic description"`
	CreatedAt *time.Time `json:"createdAt,omitempty" doc:"Created timestamp"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty" doc:"Updated timestamp"`
}

// --- Tag ---

type CreateTagRequest struct {
	Slug          string          `json:"slug" doc:"Tag slug" minLength:"1" maxLength:"100"`
	Group         entity.TagGroup `json:"group" doc:"Tag group"`
	IsMultiSelect bool            `json:"isMultiSelect" doc:"Whether multiple tags from this group can be selected together"`
	Icon          *string         `json:"icon,omitempty" doc:"Icon identifier" maxLength:"50"`
	SortOrder     int             `json:"sortOrder" doc:"Display order"`
	IsActive      bool            `json:"isActive" doc:"Whether tag is active"`
	NameEN        string          `json:"nameEn" doc:"English name" minLength:"1" maxLength:"200"`
	DescEN        *string         `json:"descEn,omitempty" doc:"English description"`
	NameAM        string          `json:"nameAm" doc:"Amharic name" minLength:"1" maxLength:"200"`
	DescAM        *string         `json:"descAm,omitempty" doc:"Amharic description"`
}

type CreateTagInput struct {
	Body CreateTagRequest
}

type CreateTagOutput struct {
	Body CreateTagResponseBody
}

type CreateTagResponseBody struct {
	ID uuid.UUID `json:"id" doc:"Created tag ID"`
}

type UpdateTagRequest struct {
	Slug          *string          `json:"slug,omitempty" doc:"Tag slug" maxLength:"100"`
	Group         *entity.TagGroup `json:"group,omitempty" doc:"Tag group"`
	IsMultiSelect *bool            `json:"isMultiSelect,omitempty" doc:"Whether multiple tags from this group can be selected together"`
	Icon          *string          `json:"icon,omitempty" doc:"Icon identifier" maxLength:"50"`
	SortOrder     *int             `json:"sortOrder,omitempty" doc:"Display order"`
	IsActive      *bool            `json:"isActive,omitempty" doc:"Whether tag is active"`
	NameEN        *string          `json:"nameEn,omitempty" doc:"English name" maxLength:"200"`
	DescEN        *string          `json:"descEn,omitempty" doc:"English description"`
	NameAM        *string          `json:"nameAm,omitempty" doc:"Amharic name" maxLength:"200"`
	DescAM        *string          `json:"descAm,omitempty" doc:"Amharic description"`
}

type UpdateTagInput struct {
	ID   uuid.UUID `path:"id" doc:"Tag ID"`
	Body UpdateTagRequest
}

type UpdateTagOutput struct {
	Body UpdateTagResponseBody
}

type UpdateTagResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type GetTagInput struct {
	ID uuid.UUID `path:"id" doc:"Tag ID"`
}

type GetTagOutput struct {
	Body TagResponse
}

type ListTagsInput struct {
	Page     int    `query:"page" doc:"Page number"`
	PageSize int    `query:"pageSize" doc:"Items per page"`
	Search   string `query:"search" doc:"Search term"`
}

type ListTagsOutput struct {
	Body ListTagsResponseBody
}

type ListTagsResponseBody struct {
	Data       []TagResponse `json:"data" doc:"Tag list"`
	Total      int64         `json:"total" doc:"Total count"`
	Page       int           `json:"page" doc:"Current page"`
	PageSize   int           `json:"pageSize" doc:"Items per page"`
	TotalPages int           `json:"totalPages" doc:"Total pages"`
}

type TagResponse struct {
	ID            uuid.UUID       `json:"id" doc:"Tag ID"`
	Slug          string          `json:"slug" doc:"Tag slug"`
	Group         entity.TagGroup `json:"group" doc:"Tag group"`
	IsMultiSelect bool            `json:"isMultiSelect" doc:"Whether multiple tags from this group can be selected together"`
	Icon          *string         `json:"icon,omitempty" doc:"Icon identifier"`
	SortOrder     int             `json:"sortOrder" doc:"Display order"`
	IsActive      bool            `json:"isActive" doc:"Whether tag is active"`
	NameEN        string          `json:"nameEn" doc:"English name"`
	DescEN        *string         `json:"descEn,omitempty" doc:"English description"`
	NameAM        string          `json:"nameAm" doc:"Amharic name"`
	DescAM        *string         `json:"descAm,omitempty" doc:"Amharic description"`
	CreatedAt     *time.Time      `json:"createdAt,omitempty" doc:"Created timestamp"`
	UpdatedAt     *time.Time      `json:"updatedAt,omitempty" doc:"Updated timestamp"`
}
