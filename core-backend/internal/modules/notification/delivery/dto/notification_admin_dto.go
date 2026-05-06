package dto

import (
	"time"

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
	TemplateGroup    string                      `json:"templateGroup" doc:"Template group"`
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
	TemplateGroup    string                      `json:"templateGroup" doc:"Template group"`
	Priority         entity.NotificationPriority `json:"priority" doc:"Default priority"`
	IsSystemManaged  bool                        `json:"isSystemManaged" doc:"Whether template is system-managed"`
	DefaultContent   map[string]interface{}      `json:"defaultContent" doc:"Multi-channel content"`
	VariablesSchema  map[string]interface{}      `json:"variablesSchema,omitempty" doc:"Template variable schema"`
	DefaultTTL       *int                        `json:"defaultTtl,omitempty" doc:"Default TTL in seconds"`
	Translations     []TranslationResponse       `json:"translations,omitempty" doc:"Template translations"`
}

// --- List Templates ---

type ListTemplatesInput struct {
	Category string `query:"category" doc:"Filter by category"`
	Page     int    `query:"page" doc:"Page number"`
	PageSize int    `query:"pageSize" doc:"Items per page"`
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
	ID               uuid.UUID               `json:"id" doc:"Template ID"`
	Name             string                  `json:"name" doc:"Template name"`
	NotificationType entity.NotificationType `json:"notificationType" doc:"Notification type"`
	TemplateGroup    string                  `json:"templateGroup" doc:"Template group"`
	IsSystemManaged  bool                    `json:"isSystemManaged" doc:"Whether template is system-managed"`
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

// --- Queue Status ---

type GetQueueStatusOutput struct {
	Body QueueStatusResponse
}

type QueueStatusResponse struct {
	Pending    int64 `json:"pending" doc:"Number of pending notifications"`
	Processing int64 `json:"processing" doc:"Number of processing notifications"`
	Delivered  int64 `json:"delivered" doc:"Number of delivered notifications"`
	Failed     int64 `json:"failed" doc:"Number of failed notifications"`
	Cancelled  int64 `json:"cancelled" doc:"Number of cancelled notifications"`
}

// --- Retry Failed ---

type RetryFailedInput struct {
	BatchSize int `query:"batchSize" doc:"Number of failed items to retry" default:"50"`
}

type RetryFailedOutput struct {
	Body RetryFailedResponseBody
}

type RetryFailedResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Mappers ---

func ToCreateTemplateInput(body CreateTemplateRequest) usecase.CreateTemplateInput {
	return usecase.CreateTemplateInput{
		Name:             body.Name,
		Description:      body.Description,
		NotificationType: body.NotificationType,
		TemplateGroup:    body.TemplateGroup,
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
		TemplateGroup:    tmpl.TemplateGroup,
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

// --- Campaigns ---

type CreateCampaignRequest struct {
	Name               string                  `json:"name" doc:"Campaign name" minLength:"1" maxLength:"200"`
	Description        *string                 `json:"description,omitempty" doc:"Campaign description"`
	CampaignType       entity.CampaignType     `json:"campaignType" doc:"Campaign type (broadcast or segmented)"`
	TargetSegment      *map[string]interface{} `json:"targetSegment,omitempty" doc:"Segment filters for segmented campaigns"`
	CampaignTemplateID uuid.UUID               `json:"campaignTemplateId" doc:"Campaign template ID"`
	ScheduledFor       *time.Time              `json:"scheduledFor,omitempty" doc:"Scheduled sending time"`
}

type CreateCampaignInput struct {
	Body CreateCampaignRequest
}

type CreateCampaignOutput struct {
	Body CreateCampaignResponseBody
}

type CreateCampaignResponseBody struct {
	ID uuid.UUID `json:"id" doc:"Created campaign ID"`
}

type GetCampaignInput struct {
	ID uuid.UUID `path:"id" doc:"Campaign ID"`
}

type GetCampaignOutput struct {
	Body CampaignDetailResponse
}

type CampaignDetailResponse struct {
	ID                 uuid.UUID                     `json:"id" doc:"Campaign ID"`
	Name               string                        `json:"name" doc:"Campaign name"`
	Description        *string                       `json:"description,omitempty" doc:"Campaign description"`
	CampaignType       entity.CampaignType           `json:"campaignType" doc:"Campaign type"`
	TargetSegment      map[string]interface{}        `json:"targetSegment,omitempty" doc:"Segment filters or resolved recipients"`
	CampaignTemplateID uuid.UUID                     `json:"campaignTemplateId" doc:"Campaign template ID"`
	CampaignTemplate   *CampaignTemplateInfoResponse `json:"campaignTemplate,omitempty" doc:"Campaign template details"`
	CreatedBy          *CampaignCreatedByResponse    `json:"createdBy" doc:"Creator account info"`
	ScheduledFor       *time.Time                    `json:"scheduledFor,omitempty" doc:"Scheduled sending time"`
	SentAt             *time.Time                    `json:"sentAt,omitempty" doc:"Actual sending time"`
	Status             entity.CampaignStatus         `json:"status" doc:"Campaign status"`
	CreatedAt          time.Time                     `json:"createdAt" doc:"Creation time"`
}

type CampaignTemplateInfoResponse struct {
	ID             uuid.UUID              `json:"id" doc:"Template ID"`
	Name           string                 `json:"name" doc:"Template name"`
	Description    *string                `json:"description,omitempty" doc:"Template description"`
	DefaultContent map[string]interface{} `json:"defaultContent" doc:"Multi-channel content"`
}

type CampaignCreatedByResponse struct {
	ID    uuid.UUID `json:"id" doc:"Account ID"`
	Name  string    `json:"name" doc:"Account name"`
	Email string    `json:"email" doc:"Account email"`
}

type ListCampaignsInput struct {
	Status   entity.CampaignStatus `query:"status" doc:"Filter by status"`
	Page     int                   `query:"page" doc:"Page number"`
	PageSize int                   `query:"pageSize" doc:"Items per page"`
}

type ListCampaignsOutput struct {
	Body ListCampaignsResponseBody
}

type ListCampaignsResponseBody struct {
	Data       []CampaignSummaryResponse `json:"data" doc:"Campaign list"`
	Total      int64                     `json:"total" doc:"Total count"`
	Page       int                       `json:"page" doc:"Current page"`
	PageSize   int                       `json:"pageSize" doc:"Items per page"`
	TotalPages int                       `json:"totalPages" doc:"Total pages"`
}

type CampaignSummaryResponse struct {
	ID           uuid.UUID             `json:"id" doc:"Campaign ID"`
	Name         string                `json:"name" doc:"Campaign name"`
	CampaignType entity.CampaignType   `json:"campaignType" doc:"Campaign type"`
	Status       entity.CampaignStatus `json:"status" doc:"Campaign status"`
	ScheduledFor *time.Time            `json:"scheduledFor,omitempty" doc:"Scheduled sending time"`
	SentAt       *time.Time            `json:"sentAt,omitempty" doc:"Actual sending time"`
	CreatedBy    uuid.UUID             `json:"createdBy" doc:"Creator account ID"`
	CreatedAt    time.Time             `json:"createdAt" doc:"Creation time"`
}

type UpdateCampaignRequest struct {
	Name          *string                 `json:"name,omitempty" doc:"Campaign name" maxLength:"200"`
	Description   *string                 `json:"description,omitempty" doc:"Campaign description"`
	TargetSegment *map[string]interface{} `json:"targetSegment,omitempty" doc:"Segment filters for segmented campaigns"`
	ScheduledFor  *time.Time              `json:"scheduledFor,omitempty" doc:"Scheduled sending time"`
}

type UpdateCampaignInput struct {
	ID   uuid.UUID `path:"id" doc:"Campaign ID"`
	Body UpdateCampaignRequest
}

type UpdateCampaignOutput struct {
	Body UpdateCampaignResponseBody
}

type UpdateCampaignResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type ScheduleCampaignInput struct {
	ID uuid.UUID `path:"id" doc:"Campaign ID"`
}

type ScheduleCampaignOutput struct {
	Body ScheduleCampaignResponseBody
}

type ScheduleCampaignResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type CancelCampaignInput struct {
	ID uuid.UUID `path:"id" doc:"Campaign ID"`
}

type CancelCampaignOutput struct {
	Body CancelCampaignResponseBody
}

type CancelCampaignResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

func ToCreateCampaignInput(createdBy uuid.UUID, body CreateCampaignRequest) usecase.CreateCampaignInput {
	return usecase.CreateCampaignInput{
		Name:               body.Name,
		Description:        body.Description,
		CampaignType:       body.CampaignType,
		TargetSegment:      body.TargetSegment,
		CampaignTemplateID: body.CampaignTemplateID,
		ScheduledFor:       body.ScheduledFor,
	}
}

func ToUpdateCampaignInput(body UpdateCampaignRequest) usecase.UpdateCampaignInput {
	return usecase.UpdateCampaignInput{
		Name:          body.Name,
		Description:   body.Description,
		TargetSegment: body.TargetSegment,
		ScheduledFor:  body.ScheduledFor,
	}
}

func ToCampaignDetailResponse(detail *usecase.CampaignDetail) CampaignDetailResponse {
	resp := CampaignDetailResponse{
		ID:                 detail.Campaign.ID,
		Name:               detail.Campaign.Name,
		Description:        detail.Campaign.Description,
		CampaignType:       detail.Campaign.CampaignType,
		CampaignTemplateID: detail.Campaign.CampaignTemplateID,
		ScheduledFor:       detail.Campaign.ScheduledFor,
		SentAt:             detail.Campaign.SentAt,
		Status:             detail.Campaign.Status,
		CreatedAt:          *detail.Campaign.CreatedAt,
	}
	if detail.Campaign.TargetSegment != nil {
		resp.TargetSegment = *detail.Campaign.TargetSegment
	}
	if detail.CampaignTemplate != nil {
		resp.CampaignTemplate = &CampaignTemplateInfoResponse{
			ID:             detail.CampaignTemplate.ID,
			Name:           detail.CampaignTemplate.Name,
			Description:    detail.CampaignTemplate.Description,
			DefaultContent: detail.CampaignTemplate.DefaultContent,
		}
	}
	resp.CreatedBy = &CampaignCreatedByResponse{
		ID:    detail.Campaign.CreatedBy,
		Name:  detail.CreatedByName,
		Email: detail.CreatedByEmail,
	}
	return resp
}

func ToCampaignSummaryResponse(c *entity.NotificationCampaign) CampaignSummaryResponse {
	return CampaignSummaryResponse{
		ID:           c.ID,
		Name:         c.Name,
		CampaignType: c.CampaignType,
		Status:       c.Status,
		ScheduledFor: c.ScheduledFor,
		SentAt:       c.SentAt,
		CreatedBy:    c.CreatedBy,
	}
}
