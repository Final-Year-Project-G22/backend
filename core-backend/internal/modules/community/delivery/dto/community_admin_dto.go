package dto

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type CreateCommunityCategoryRequest struct {
	Name             string     `json:"name" doc:"Category name"`
	Slug             string     `json:"slug" doc:"Category slug"`
	Description      *string    `json:"description,omitempty" doc:"Category description"`
	ParentCategoryID *uuid.UUID `json:"parentCategoryId,omitempty" doc:"Parent category ID"`
	IsActive         bool       `json:"isActive" doc:"Active flag"`
}

type CreateCommunityCategoryInput struct {
	Body CreateCommunityCategoryRequest
}

type CreateCommunityCategoryOutput struct {
	Body CreateCommunityCategoryResponseBody
}

type CreateCommunityCategoryResponseBody struct {
	ID uuid.UUID `json:"id" doc:"Created category ID"`
}

type UpdateCommunityCategoryRequest struct {
	Name             *string    `json:"name,omitempty" doc:"Category name"`
	Slug             *string    `json:"slug,omitempty" doc:"Category slug"`
	Description      *string    `json:"description,omitempty" doc:"Category description"`
	ParentCategoryID *uuid.UUID `json:"parentCategoryId,omitempty" doc:"Parent category ID"`
	IsActive         *bool      `json:"isActive,omitempty" doc:"Active flag"`
}

type UpdateCommunityCategoryInput struct {
	ID   uuid.UUID `path:"id" doc:"Category ID"`
	Body UpdateCommunityCategoryRequest
}

type UpdateCommunityCategoryOutput struct {
	Body UpdateCommunityCategoryResponseBody
}

type UpdateCommunityCategoryResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type DeleteCommunityCategoryInput struct {
	ID uuid.UUID `path:"id" doc:"Category ID"`
}

type DeleteCommunityCategoryOutput struct {
	Body DeleteCommunityCategoryResponseBody
}

type DeleteCommunityCategoryResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type AdminListCommunityCategoriesInput struct {
	Page            int  `query:"page" doc:"Page number"`
	PageSize        int  `query:"pageSize" doc:"Page size"`
	IncludeInactive bool `query:"includeInactive" doc:"Include inactive categories"`
}

type AdminListCommunityCategoriesOutput struct {
	Body ListCategoriesResponseBody
}

func ToCreateCategoryInput(input CreateCommunityCategoryRequest) usecase.CreateCategoryInput {
	return usecase.CreateCategoryInput{
		Name:             input.Name,
		Slug:             input.Slug,
		Description:      input.Description,
		ParentCategoryID: input.ParentCategoryID,
		IsActive:         input.IsActive,
	}
}

func ToUpdateCategoryInput(input UpdateCommunityCategoryRequest) usecase.UpdateCategoryInput {
	return usecase.UpdateCategoryInput{
		Name:             input.Name,
		Slug:             input.Slug,
		Description:      input.Description,
		ParentCategoryID: input.ParentCategoryID,
		IsActive:         input.IsActive,
	}
}

func ToAdminQueryOptions(page, pageSize int) query.QueryOptions {
	return ToQueryOptions(page, pageSize)
}
