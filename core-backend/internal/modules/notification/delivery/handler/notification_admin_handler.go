package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/i18n"
)

type NotificationAdminHandler struct {
	templateUC usecase.NotificationTemplateUsecase
	deliveryUC usecase.NotificationDeliveryUsecase
	campaignUC usecase.NotificationCampaignUsecase
	queueRepo  repository.NotificationQueueRepository
}

func NewNotificationAdminHandler(
	templateUC usecase.NotificationTemplateUsecase,
	deliveryUC usecase.NotificationDeliveryUsecase,
	campaignUC usecase.NotificationCampaignUsecase,
	queueRepo repository.NotificationQueueRepository,
) *NotificationAdminHandler {
	return &NotificationAdminHandler{
		templateUC: templateUC,
		deliveryUC: deliveryUC,
		campaignUC: campaignUC,
		queueRepo:  queueRepo,
	}
}

func (h *NotificationAdminHandler) HandleCreateTemplate(ctx context.Context, input *dto.CreateTemplateInput) (*dto.CreateTemplateOutput, error) {
	result, err := h.templateUC.CreateTemplate(ctx, dto.ToCreateTemplateInput(input.Body))
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.CreateTemplateOutput{Body: dto.CreateTemplateResponseBody{ID: result.ID}}, nil
}

func (h *NotificationAdminHandler) HandleGetTemplate(ctx context.Context, input *dto.GetTemplateInput) (*dto.GetTemplateOutput, error) {
	tmpl, err := h.templateUC.GetTemplate(ctx, input.ID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	translations, err := h.templateUC.GetTranslations(ctx, input.ID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.GetTemplateOutput{Body: dto.ToTemplateDetailResponse(tmpl, translations)}, nil
}

func (h *NotificationAdminHandler) HandleListTemplates(ctx context.Context, input *dto.ListTemplatesInput) (*dto.ListTemplatesOutput, error) {
	q := dto.ToQueryOptions(input.Page, input.PageSize)
	templates, err := h.templateUC.ListTemplates(ctx, input.TemplateGroup, q)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	data := make([]dto.TemplateSummaryResponse, 0, len(templates))
	for _, tmpl := range templates {
		data = append(data, dto.TemplateSummaryResponse{
			ID:               tmpl.ID,
			Name:             tmpl.Name,
			NotificationType: tmpl.NotificationType,
			TemplateGroup:    tmpl.TemplateGroup,
			IsSystemManaged:  tmpl.IsSystemManaged,
		})
	}

	return &dto.ListTemplatesOutput{Body: dto.ListTemplatesResponseBody{
		Data:       data,
		Total:      int64(len(data)),
		Page:       q.Page,
		PageSize:   q.PageSize,
		TotalPages: int((len(data) + q.PageSize - 1) / q.PageSize),
	}}, nil
}

func (h *NotificationAdminHandler) HandleUpdateTemplate(ctx context.Context, input *dto.UpdateTemplateInput) (*dto.UpdateTemplateOutput, error) {
	if _, err := h.templateUC.UpdateTemplate(ctx, input.ID, dto.ToUpdateTemplateInput(input.Body)); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UpdateTemplateOutput{Body: dto.UpdateTemplateResponseBody{Message: i18n.T(ctx, "notification.successes.adminTemplateUpdated")}}, nil
}

func (h *NotificationAdminHandler) HandleDeleteTemplate(ctx context.Context, input *dto.DeleteTemplateInput) (*dto.DeleteTemplateOutput, error) {
	if err := h.templateUC.DeleteTemplate(ctx, input.ID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.DeleteTemplateOutput{Body: dto.DeleteTemplateResponseBody{Message: i18n.T(ctx, "notification.successes.adminTemplateDeleted")}}, nil
}

func (h *NotificationAdminHandler) HandleAddTranslation(ctx context.Context, input *dto.AddTranslationInput) (*dto.AddTranslationOutput, error) {
	ucInput := dto.ToAddTranslationInput(input.Body)
	ucInput.TemplateID = input.TemplateID
	result, err := h.templateUC.AddTranslation(ctx, ucInput)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.AddTranslationOutput{Body: dto.AddTranslationResponseBody{ID: result.ID}}, nil
}

func (h *NotificationAdminHandler) HandleUpdateTranslation(ctx context.Context, input *dto.UpdateTranslationInput) (*dto.UpdateTranslationOutput, error) {
	if _, err := h.templateUC.UpdateTranslation(ctx, input.TemplateID, input.Language, dto.ToUpdateTranslationInput(input.Body)); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UpdateTranslationOutput{Body: dto.UpdateTranslationResponseBody{Message: i18n.T(ctx, "notification.successes.adminTranslationUpdated")}}, nil
}

func (h *NotificationAdminHandler) HandleDeleteTranslation(ctx context.Context, input *dto.DeleteTranslationInput) (*dto.DeleteTranslationOutput, error) {
	if err := h.templateUC.DeleteTranslation(ctx, input.TemplateID, input.Language); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.DeleteTranslationOutput{Body: dto.DeleteTranslationResponseBody{Message: i18n.T(ctx, "notification.successes.adminTranslationDeleted")}}, nil
}

// --- Monitoring ---

func (h *NotificationAdminHandler) HandleGetQueueStatus(ctx context.Context, input *struct{}) (*dto.GetQueueStatusOutput, error) {
	pending, err := h.queueRepo.CountByStatus(ctx, entity.NotificationStatusPending)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	processing, err := h.queueRepo.CountByStatus(ctx, entity.NotificationStatusProcessing)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	delivered, err := h.queueRepo.CountByStatus(ctx, entity.NotificationStatusDelivered)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	failed, err := h.queueRepo.CountByStatus(ctx, entity.NotificationStatusFailed)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	cancelled, err := h.queueRepo.CountByStatus(ctx, entity.NotificationStatusCancelled)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.GetQueueStatusOutput{Body: dto.QueueStatusResponse{
		Pending:    pending,
		Processing: processing,
		Delivered:  delivered,
		Failed:     failed,
		Cancelled:  cancelled,
	}}, nil
}

func (h *NotificationAdminHandler) HandleRetryFailed(ctx context.Context, input *dto.RetryFailedInput) (*dto.RetryFailedOutput, error) {
	batchSize := input.BatchSize
	if batchSize <= 0 {
		batchSize = 50
	}
	if err := h.deliveryUC.RetryFailed(ctx, batchSize); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.RetryFailedOutput{Body: dto.RetryFailedResponseBody{Message: i18n.T(ctx, "notification.successes.retryInitiated")}}, nil
}

// --- Campaigns ---

func (h *NotificationAdminHandler) HandleCreateCampaign(ctx context.Context, input *dto.CreateCampaignInput) (*dto.CreateCampaignOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	result, err := h.campaignUC.CreateCampaign(ctx, accountID, dto.ToCreateCampaignInput(accountID, input.Body))
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.CreateCampaignOutput{Body: dto.CreateCampaignResponseBody{ID: result.ID}}, nil
}

func (h *NotificationAdminHandler) HandleGetCampaign(ctx context.Context, input *dto.GetCampaignInput) (*dto.GetCampaignOutput, error) {
	detail, err := h.campaignUC.GetCampaign(ctx, input.ID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.GetCampaignOutput{Body: dto.ToCampaignDetailResponse(detail)}, nil
}

func (h *NotificationAdminHandler) HandleListCampaigns(ctx context.Context, input *dto.ListCampaignsInput) (*dto.ListCampaignsOutput, error) {
	q := dto.ToQueryOptions(input.Page, input.PageSize)
	var statusFilter *entity.CampaignStatus
	if input.Status != "" {
		s := input.Status
		statusFilter = &s
	}
	items, total, err := h.campaignUC.ListCampaigns(ctx, statusFilter, q)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	data := make([]dto.CampaignSummaryResponse, len(items))
	for i, c := range items {
		data[i] = dto.ToCampaignSummaryResponse(c)
	}
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	return &dto.ListCampaignsOutput{Body: dto.ListCampaignsResponseBody{
		Data:       data,
		Total:      total,
		Page:       q.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}}, nil
}

func (h *NotificationAdminHandler) HandleUpdateCampaign(ctx context.Context, input *dto.UpdateCampaignInput) (*dto.UpdateCampaignOutput, error) {
	if _, err := h.campaignUC.UpdateCampaign(ctx, input.ID, dto.ToUpdateCampaignInput(input.Body)); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UpdateCampaignOutput{Body: dto.UpdateCampaignResponseBody{Message: i18n.T(ctx, "notification.successes.campaignUpdated")}}, nil
}

func (h *NotificationAdminHandler) HandleScheduleCampaign(ctx context.Context, input *dto.ScheduleCampaignInput) (*dto.ScheduleCampaignOutput, error) {
	if err := h.campaignUC.ScheduleCampaign(ctx, usecase.ScheduleCampaignInput{CampaignID: input.ID}); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.ScheduleCampaignOutput{Body: dto.ScheduleCampaignResponseBody{Message: i18n.T(ctx, "notification.successes.campaignScheduled")}}, nil
}

func (h *NotificationAdminHandler) HandleCancelCampaign(ctx context.Context, input *dto.CancelCampaignInput) (*dto.CancelCampaignOutput, error) {
	if err := h.campaignUC.CancelCampaign(ctx, input.ID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.CancelCampaignOutput{Body: dto.CancelCampaignResponseBody{Message: i18n.T(ctx, "notification.successes.campaignCancelled")}}, nil
}
