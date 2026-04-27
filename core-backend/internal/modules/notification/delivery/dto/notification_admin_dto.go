package dto

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

// --- Create Template ---

type CreateTemplateRequest struct {
	Name             string                      `json:"name" doc:"Template name" minLength:"1" maxLength:"200"`
	Description      *string                     `json:"description,omitempty" doc:"Template description"`
	NotificationType entity.NotificationType     `json:"notificationType" doc:"Notification type"`
	Category         entity.NotificationCategory `json:"category" doc:"Notification category"`
	Priority         entity.NotificationPriority `json:"priority" doc:"Default priority"`
	DefaultContent   map[string]interface{}      `json:"defaultContent" doc:"Multi-channel content"`
	VariablesSchema  *map[string]interface{}     `json:"variablesSchema,omitempty" doc:"Template variable schema"`
	DefaultTTL       *int                        `json:"defaultTtl,omitempty" doc:"Default TTL in seconds"`
}

type CreateTemplateInput struct {
	Body CreateTemplateRequest
}

type CreateTemplateOutput struct {
	Body CreateTemplateResponseBody
}

type CreateTemplateResponseBody struct {
	ID uuid.UUID `json:"id" doc:"Created template ID"`
}

// --- Get Template ---

type GetTemplateInput struct {
	ID uuid.UUID `path:"id" doc:"Template ID"`
}

type GetTemplateOutput struct {
	Body TemplateDetailResponse
}

type TemplateDetailResponse struct {
	ID               uuid.UUID                   `json:"id" doc:"Template ID"`
	Name             string                      `json:"name" doc:"Template name"`
	Description      *string                     `json:"description,omitempty" doc:"Template description"`
	NotificationType entity.NotificationType     `json:"notificationType" doc:"Notification type"`
	Category         entity.NotificationCategory `json:"category" doc:"Notification category"`
	Priority         entity.NotificationPriority `json:"priority" doc:"Default priority"`
	IsSystemManaged  bool                        `json:"isSystemManaged" doc:"Whether template is system-managed"`
	DefaultContent   map[string]interface{}      `json:"defaultContent" doc:"Multi-channel content"`
	VariablesSchema  map[string]interface{}      `json:"variablesSchema,omitempty" doc:"Template variable schema"`
	DefaultTTL       *int                        `json:"defaultTtl,omitempty" doc:"Default TTL in seconds"`
	Translations     []TranslationResponse       `json:"translations,omitempty" doc:"Template translations"`
}

// --- List Templates ---

type ListTemplatesInput struct {
	Category *entity.NotificationCategory `query:"category" doc:"Filter by category"`
	Page     int                          `query:"page" doc:"Page number"`
	PageSize int                          `query:"pageSize" doc:"Items per page"`
}

type ListTemplatesOutput struct {
	Body ListTemplatesResponseBody
}

type ListTemplatesResponseBody struct {
	Data       []TemplateSummaryResponse `json:"data" doc:"Template list"`
	Total      int64                     `json:"total" doc:"Total count"`
	Page       int                       `json:"page" doc:"Current page"`
	PageSize   int                       `json:"pageSize" doc:"Items per page"`
	TotalPages int                       `json:"totalPages" doc:"Total pages"`
}

type TemplateSummaryResponse struct {
	ID               uuid.UUID                   `json:"id" doc:"Template ID"`
	Name             string                      `json:"name" doc:"Template name"`
	NotificationType entity.NotificationType     `json:"notificationType" doc:"Notification type"`
	Category         entity.NotificationCategory `json:"category" doc:"Notification category"`
	IsSystemManaged  bool                        `json:"isSystemManaged" doc:"Whether template is system-managed"`
}

// --- Update Template ---

type UpdateTemplateRequest struct {
	Name            *string                      `json:"name,omitempty" doc:"Template name" maxLength:"200"`
	Description     *string                      `json:"description,omitempty" doc:"Template description"`
	Priority        *entity.NotificationPriority `json:"priority,omitempty" doc:"Default priority"`
	DefaultContent  *map[string]interface{}      `json:"defaultContent,omitempty" doc:"Multi-channel content"`
	VariablesSchema *map[string]interface{}      `json:"variablesSchema,omitempty" doc:"Template variable schema"`
	DefaultTTL      *int                         `json:"defaultTtl,omitempty" doc:"Default TTL in seconds"`
}

type UpdateTemplateInput struct {
	ID   uuid.UUID `path:"id" doc:"Template ID"`
	Body UpdateTemplateRequest
}

type UpdateTemplateOutput struct {
	Body UpdateTemplateResponseBody
}

type UpdateTemplateResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Delete Template ---

type DeleteTemplateInput struct {
	ID uuid.UUID `path:"id" doc:"Template ID"`
}

type DeleteTemplateOutput struct {
	Body DeleteTemplateResponseBody
}

type DeleteTemplateResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Translations ---

type AddTranslationRequest struct {
	Language string                 `json:"language" doc:"Language code" minLength:"2" maxLength:"10"`
	Subject  string                 `json:"subject" doc:"Localized subject" minLength:"1" maxLength:"500"`
	Content  map[string]interface{} `json:"content" doc:"Multi-channel localized content"`
}

type AddTranslationInput struct {
	TemplateID uuid.UUID `path:"id" doc:"Template ID"`
	Body       AddTranslationRequest
}

type AddTranslationOutput struct {
	Body AddTranslationResponseBody
}

type AddTranslationResponseBody struct {
	ID uuid.UUID `json:"id" doc:"Created translation ID"`
}

type UpdateTranslationRequest struct {
	Subject *string                 `json:"subject,omitempty" doc:"Localized subject" maxLength:"500"`
	Content *map[string]interface{} `json:"content,omitempty" doc:"Multi-channel localized content"`
}

type UpdateTranslationInput struct {
	TemplateID uuid.UUID `path:"id" doc:"Template ID"`
	Language   string    `path:"lang" doc:"Language code"`
	Body       UpdateTranslationRequest
}

type UpdateTranslationOutput struct {
	Body UpdateTranslationResponseBody
}

type UpdateTranslationResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type DeleteTranslationInput struct {
	TemplateID uuid.UUID `path:"id" doc:"Template ID"`
	Language   string    `path:"lang" doc:"Language code"`
}

type DeleteTranslationOutput struct {
	Body DeleteTranslationResponseBody
}

type DeleteTranslationResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type TranslationResponse struct {
	ID       uuid.UUID              `json:"id" doc:"Translation ID"`
	Language string                 `json:"language" doc:"Language code"`
	Subject  string                 `json:"subject" doc:"Localized subject"`
	Content  map[string]interface{} `json:"content" doc:"Multi-channel localized content"`
}

// --- Mappers ---

func ToCreateTemplateInput(body CreateTemplateRequest) usecase.CreateTemplateInput {
	return usecase.CreateTemplateInput{
		Name:             body.Name,
		Description:      body.Description,
		NotificationType: body.NotificationType,
		Category:         body.Category,
		Priority:         body.Priority,
		DefaultContent:   body.DefaultContent,
		VariablesSchema:  body.VariablesSchema,
		DefaultTTL:       body.DefaultTTL,
	}
}

func ToUpdateTemplateInput(body UpdateTemplateRequest) usecase.UpdateTemplateInput {
	return usecase.UpdateTemplateInput{
		Name:            body.Name,
		Description:     body.Description,
		Priority:        body.Priority,
		DefaultContent:  body.DefaultContent,
		VariablesSchema: body.VariablesSchema,
		DefaultTTL:      body.DefaultTTL,
	}
}

func ToAddTranslationInput(body AddTranslationRequest) usecase.CreateTemplateTranslationInput {
	return usecase.CreateTemplateTranslationInput{
		Language: body.Language,
		Subject:  body.Subject,
		Content:  body.Content,
	}
}

func ToUpdateTranslationInput(body UpdateTranslationRequest) usecase.UpdateTemplateTranslationInput {
	return usecase.UpdateTemplateTranslationInput{
		Subject: body.Subject,
		Content: body.Content,
	}
}

func ToQueryOptions(page, pageSize int) query.QueryOptions {
	q := query.DefaultQueryOptions()
	if page > 0 {
		q.Page = page
	}
	if pageSize > 0 {
		q.PageSize = pageSize
	}
	return q
}

func ToTemplateDetailResponse(tmpl *entity.NotificationTemplate, translations []*entity.NotificationTemplateTranslation) TemplateDetailResponse {
	resp := TemplateDetailResponse{
		ID:               tmpl.ID,
		Name:             tmpl.Name,
		Description:      tmpl.Description,
		NotificationType: tmpl.NotificationType,
		Category:         tmpl.Category,
		Priority:         tmpl.Priority,
		IsSystemManaged:  tmpl.IsSystemManaged,
		DefaultContent:   tmpl.DefaultContent,
		DefaultTTL:       tmpl.DefaultTTL,
	}
	if tmpl.VariablesSchema != nil {
		resp.VariablesSchema = *tmpl.VariablesSchema
	}
	if len(translations) > 0 {
		resp.Translations = make([]TranslationResponse, 0, len(translations))
		for _, t := range translations {
			resp.Translations = append(resp.Translations, TranslationResponse{
				ID:       t.ID,
				Language: t.Language,
				Subject:  t.Subject,
				Content:  t.Content,
			})
		}
	}
	return resp
}
