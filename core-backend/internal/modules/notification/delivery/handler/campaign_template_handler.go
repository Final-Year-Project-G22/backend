package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
)

type CampaignTemplateHandler struct {
	campaignTemplateUC usecase.CampaignTemplateUsecase
}

func NewCampaignTemplateHandler(campaignTemplateUC usecase.CampaignTemplateUsecase) *CampaignTemplateHandler {
	return &CampaignTemplateHandler{campaignTemplateUC: campaignTemplateUC}
}

func (h *CampaignTemplateHandler) HandleCreateCampaignTemplate(ctx context.Context, input *dto.CreateCampaignTemplateInput) (*dto.CreateCampaignTemplateOutput, error) {
	result, err := h.campaignTemplateUC.Create(ctx, dto.ToCreateCampaignTemplateInput(input.Body))
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.CreateCampaignTemplateOutput{Body: dto.CreateCampaignTemplateResponseBody{ID: result.ID}}, nil
}

func (h *CampaignTemplateHandler) HandleGetCampaignTemplate(ctx context.Context, input *dto.GetCampaignTemplateInput) (*dto.GetCampaignTemplateOutput, error) {
	tmpl, err := h.campaignTemplateUC.GetByID(ctx, input.ID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	translations, err := h.campaignTemplateUC.GetTranslations(ctx, input.ID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.GetCampaignTemplateOutput{Body: dto.ToCampaignTemplateDetailResponse(tmpl, translations)}, nil
}

func (h *CampaignTemplateHandler) HandleListCampaignTemplates(ctx context.Context, input *dto.ListCampaignTemplatesInput) (*dto.ListCampaignTemplatesOutput, error) {
	q := dto.ToQueryOptions(input.Page, input.PageSize)
	templates, total, err := h.campaignTemplateUC.List(ctx, q)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	data := make([]dto.CampaignTemplateSummaryResponse, 0, len(templates))
	for _, tmpl := range templates {
		data = append(data, dto.ToCampaignTemplateSummaryResponse(tmpl))
	}

	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &dto.ListCampaignTemplatesOutput{Body: dto.ListCampaignTemplatesResponseBody{
		Data:       data,
		Total:      total,
		Page:       q.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}}, nil
}

func (h *CampaignTemplateHandler) HandleUpdateCampaignTemplate(ctx context.Context, input *dto.UpdateCampaignTemplateInput) (*dto.UpdateCampaignTemplateOutput, error) {
	if err := h.campaignTemplateUC.Update(ctx, input.ID, dto.ToUpdateCampaignTemplateInput(input.Body)); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UpdateCampaignTemplateOutput{Body: dto.UpdateCampaignTemplateResponseBody{Message: "Campaign template updated"}}, nil
}

func (h *CampaignTemplateHandler) HandleDeleteCampaignTemplate(ctx context.Context, input *dto.DeleteCampaignTemplateInput) (*dto.DeleteCampaignTemplateOutput, error) {
	if err := h.campaignTemplateUC.Delete(ctx, input.ID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.DeleteCampaignTemplateOutput{Body: dto.DeleteCampaignTemplateResponseBody{Message: "Campaign template deleted"}}, nil
}

func (h *CampaignTemplateHandler) HandleAddCampaignTemplateTranslation(ctx context.Context, input *dto.AddCampaignTemplateTranslationInput) (*dto.AddCampaignTemplateTranslationOutput, error) {
	ucInput := dto.ToAddCampaignTranslationInput(input.Body)
	ucInput.TemplateID = input.TemplateID
	result, err := h.campaignTemplateUC.AddTranslation(ctx, ucInput)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.AddCampaignTemplateTranslationOutput{Body: dto.AddCampaignTemplateTranslationResponseBody{ID: result.ID}}, nil
}

func (h *CampaignTemplateHandler) HandleUpdateCampaignTemplateTranslation(ctx context.Context, input *dto.UpdateCampaignTemplateTranslationInput) (*dto.UpdateCampaignTemplateTranslationOutput, error) {
	if _, err := h.campaignTemplateUC.UpdateTranslation(ctx, input.TemplateID, input.Language, dto.ToUpdateCampaignTranslationInput(input.Body)); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UpdateCampaignTemplateTranslationOutput{Body: dto.UpdateCampaignTemplateTranslationResponseBody{Message: "Translation updated"}}, nil
}

func (h *CampaignTemplateHandler) HandleDeleteCampaignTemplateTranslation(ctx context.Context, input *dto.DeleteCampaignTemplateTranslationInput) (*dto.DeleteCampaignTemplateTranslationOutput, error) {
	if err := h.campaignTemplateUC.DeleteTranslation(ctx, input.TemplateID, input.Language); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.DeleteCampaignTemplateTranslationOutput{Body: dto.DeleteCampaignTemplateTranslationResponseBody{Message: "Translation deleted"}}, nil
}

func (h *CampaignTemplateHandler) HandleListCampaignTemplatesWith(ctx context.Context, input *dto.ListCampaignTemplatesInput) (*dto.ListCampaignTemplatesOutput, error) {
	return h.HandleListCampaignTemplates(ctx, input)
}
