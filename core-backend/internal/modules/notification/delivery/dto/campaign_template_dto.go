package dto

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	"github.com/google/uuid"
)

// --- Create Campaign Template ---

type CreateCampaignTemplateRequest struct {
	Name           string                 `json:"name" doc:"Template name" minLength:"1" maxLength:"200"`
	Description    *string                `json:"description,omitempty" doc:"Template description"`
	DefaultContent map[string]interface{} `json:"defaultContent" doc:"Multi-channel content (in_app, email, push, sms)"`
}

type CreateCampaignTemplateInput struct {
	Body CreateCampaignTemplateRequest
}

type CreateCampaignTemplateOutput struct {
	Body CreateCampaignTemplateResponseBody
}

type CreateCampaignTemplateResponseBody struct {
	ID uuid.UUID `json:"id" doc:"Created template ID"`
}

// --- Get Campaign Template ---

type GetCampaignTemplateInput struct {
	ID uuid.UUID `path:"id" doc:"Campaign template ID"`
}

type GetCampaignTemplateOutput struct {
	Body CampaignTemplateDetailResponse
}

// --- List Campaign Templates ---

type ListCampaignTemplatesInput struct {
	Page     int `query:"page" doc:"Page number"`
	PageSize int `query:"pageSize" doc:"Items per page"`
}

type ListCampaignTemplatesOutput struct {
	Body ListCampaignTemplatesResponseBody
}

type ListCampaignTemplatesResponseBody struct {
	Data       []CampaignTemplateSummaryResponse `json:"data" doc:"Template list"`
	Total      int64                             `json:"total" doc:"Total count"`
	Page       int                               `json:"page" doc:"Current page"`
	PageSize   int                               `json:"pageSize" doc:"Items per page"`
	TotalPages int                               `json:"totalPages" doc:"Total pages"`
}

// --- Update Campaign Template ---

type UpdateCampaignTemplateRequest struct {
	Name           *string                 `json:"name,omitempty" doc:"Template name" maxLength:"200"`
	Description    *string                 `json:"description,omitempty" doc:"Template description"`
	DefaultContent *map[string]interface{} `json:"defaultContent,omitempty" doc:"Multi-channel content"`
}

type UpdateCampaignTemplateInput struct {
	ID   uuid.UUID `path:"id" doc:"Campaign template ID"`
	Body UpdateCampaignTemplateRequest
}

type UpdateCampaignTemplateOutput struct {
	Body UpdateCampaignTemplateResponseBody
}

type UpdateCampaignTemplateResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Delete Campaign Template ---

type DeleteCampaignTemplateInput struct {
	ID uuid.UUID `path:"id" doc:"Campaign template ID"`
}

type DeleteCampaignTemplateOutput struct {
	Body DeleteCampaignTemplateResponseBody
}

type DeleteCampaignTemplateResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Translation Add ---

type AddCampaignTemplateTranslationRequest struct {
	Language string                 `json:"language" doc:"Language code" minLength:"2" maxLength:"10"`
	Content  map[string]interface{} `json:"content" doc:"Multi-channel localized content"`
}

type AddCampaignTemplateTranslationInput struct {
	TemplateID uuid.UUID `path:"id" doc:"Campaign template ID"`
	Body       AddCampaignTemplateTranslationRequest
}

type AddCampaignTemplateTranslationOutput struct {
	Body AddCampaignTemplateTranslationResponseBody
}

type AddCampaignTemplateTranslationResponseBody struct {
	ID uuid.UUID `json:"id" doc:"Created translation ID"`
}

// --- Translation Update ---

type UpdateCampaignTemplateTranslationRequest struct {
	Content *map[string]interface{} `json:"content,omitempty" doc:"Multi-channel localized content"`
}

type UpdateCampaignTemplateTranslationInput struct {
	TemplateID uuid.UUID `path:"id" doc:"Campaign template ID"`
	Language   string    `path:"lang" doc:"Language code"`
	Body       UpdateCampaignTemplateTranslationRequest
}

type UpdateCampaignTemplateTranslationOutput struct {
	Body UpdateCampaignTemplateTranslationResponseBody
}

type UpdateCampaignTemplateTranslationResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Translation Delete ---

type DeleteCampaignTemplateTranslationInput struct {
	TemplateID uuid.UUID `path:"id" doc:"Campaign template ID"`
	Language   string    `path:"lang" doc:"Language code"`
}

type DeleteCampaignTemplateTranslationOutput struct {
	Body DeleteCampaignTemplateTranslationResponseBody
}

type DeleteCampaignTemplateTranslationResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Response Types ---

type CampaignTemplateDetailResponse struct {
	ID             uuid.UUID                             `json:"id" doc:"Template ID"`
	Name           string                                `json:"name" doc:"Template name"`
	Description    *string                               `json:"description,omitempty" doc:"Template description"`
	DefaultContent map[string]interface{}                `json:"defaultContent" doc:"Multi-channel content"`
	Translations   []CampaignTemplateTranslationResponse `json:"translations,omitempty" doc:"Template translations"`
	CreatedAt      time.Time                             `json:"createdAt" doc:"Creation time"`
}

type CampaignTemplateSummaryResponse struct {
	ID          uuid.UUID `json:"id" doc:"Template ID"`
	Name        string    `json:"name" doc:"Template name"`
	Description *string   `json:"description,omitempty" doc:"Template description"`
	CreatedAt   time.Time `json:"createdAt" doc:"Creation time"`
}

type CampaignTemplateTranslationResponse struct {
	ID       uuid.UUID              `json:"id" doc:"Translation ID"`
	Language string                 `json:"language" doc:"Language code"`
	Content  map[string]interface{} `json:"content" doc:"Multi-channel localized content"`
}

// --- Mappers ---

func ToCreateCampaignTemplateInput(body CreateCampaignTemplateRequest) usecase.CreateCampaignTemplateInput {
	return usecase.CreateCampaignTemplateInput{
		Name:           body.Name,
		Description:    body.Description,
		DefaultContent: body.DefaultContent,
	}
}

func ToUpdateCampaignTemplateInput(body UpdateCampaignTemplateRequest) usecase.UpdateCampaignTemplateInput {
	return usecase.UpdateCampaignTemplateInput{
		Name:           body.Name,
		Description:    body.Description,
		DefaultContent: body.DefaultContent,
	}
}

func ToAddCampaignTranslationInput(body AddCampaignTemplateTranslationRequest) usecase.CreateCampaignTemplateTranslationInput {
	return usecase.CreateCampaignTemplateTranslationInput{
		Language: body.Language,
		Content:  body.Content,
	}
}

func ToUpdateCampaignTranslationInput(body UpdateCampaignTemplateTranslationRequest) usecase.UpdateCampaignTemplateTranslationInput {
	return usecase.UpdateCampaignTemplateTranslationInput{
		Content: body.Content,
	}
}

func ToCampaignTemplateDetailResponse(tmpl *entity.CampaignTemplate, translations []*entity.CampaignTemplateTranslation) CampaignTemplateDetailResponse {
	resp := CampaignTemplateDetailResponse{
		ID:             tmpl.ID,
		Name:           tmpl.Name,
		Description:    tmpl.Description,
		DefaultContent: tmpl.DefaultContent,
		CreatedAt:      *tmpl.CreatedAt,
	}
	if len(translations) > 0 {
		resp.Translations = make([]CampaignTemplateTranslationResponse, 0, len(translations))
		for _, t := range translations {
			resp.Translations = append(resp.Translations, CampaignTemplateTranslationResponse{
				ID:       t.ID,
				Language: t.Language,
				Content:  t.Content,
			})
		}
	}
	return resp
}

func ToCampaignTemplateSummaryResponse(tmpl *entity.CampaignTemplate) CampaignTemplateSummaryResponse {
	return CampaignTemplateSummaryResponse{
		ID:          tmpl.ID,
		Name:        tmpl.Name,
		Description: tmpl.Description,
		CreatedAt:   *tmpl.CreatedAt,
	}
}
