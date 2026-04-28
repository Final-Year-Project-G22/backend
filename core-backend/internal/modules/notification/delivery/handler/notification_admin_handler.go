package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
)

type NotificationAdminHandler struct {
	templateUC usecase.NotificationTemplateUsecase
	deliveryUC usecase.NotificationDeliveryUsecase
	queueRepo  repository.NotificationQueueRepository
}

func NewNotificationAdminHandler(
	templateUC usecase.NotificationTemplateUsecase,
	deliveryUC usecase.NotificationDeliveryUsecase,
	queueRepo repository.NotificationQueueRepository,
) *NotificationAdminHandler {
	return &NotificationAdminHandler{
		templateUC: templateUC,
		deliveryUC: deliveryUC,
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
	var category *entity.NotificationCategory
	if input.Category != "" {
		cat := entity.NotificationCategory(input.Category)
		category = &cat
	}
	templates, err := h.templateUC.ListTemplates(ctx, category, q)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	data := make([]dto.TemplateSummaryResponse, 0, len(templates))
	for _, tmpl := range templates {
		data = append(data, dto.TemplateSummaryResponse{
			ID:               tmpl.ID,
			Name:             tmpl.Name,
			NotificationType: tmpl.NotificationType,
			Category:         tmpl.Category,
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
	return &dto.UpdateTemplateOutput{Body: dto.UpdateTemplateResponseBody{Message: "Template updated"}}, nil
}

func (h *NotificationAdminHandler) HandleDeleteTemplate(ctx context.Context, input *dto.DeleteTemplateInput) (*dto.DeleteTemplateOutput, error) {
	if err := h.templateUC.DeleteTemplate(ctx, input.ID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.DeleteTemplateOutput{Body: dto.DeleteTemplateResponseBody{Message: "Template deleted"}}, nil
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
	return &dto.UpdateTranslationOutput{Body: dto.UpdateTranslationResponseBody{Message: "Translation updated"}}, nil
}

func (h *NotificationAdminHandler) HandleDeleteTranslation(ctx context.Context, input *dto.DeleteTranslationInput) (*dto.DeleteTranslationOutput, error) {
	if err := h.templateUC.DeleteTranslation(ctx, input.TemplateID, input.Language); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.DeleteTranslationOutput{Body: dto.DeleteTranslationResponseBody{Message: "Translation deleted"}}, nil
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
	return &dto.RetryFailedOutput{Body: dto.RetryFailedResponseBody{Message: "Retry initiated"}}, nil
}
