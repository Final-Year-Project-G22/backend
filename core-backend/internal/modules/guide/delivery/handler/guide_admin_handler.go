package handler

import (
	"context"
	"errors"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/delivery/dto"
	guideerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
)

type GuideAdminHandler struct {
	guideAdminUC   usecase.GuideManagementUseCase
	journeyAdminUC usecase.JourneyManagementUseCase
}

func NewGuideAdminHandler(guideAdminUC usecase.GuideManagementUseCase, journeyAdminUC usecase.JourneyManagementUseCase) *GuideAdminHandler {
	return &GuideAdminHandler{
		guideAdminUC:   guideAdminUC,
		journeyAdminUC: journeyAdminUC,
	}
}

func (h *GuideAdminHandler) HandleListGuides(ctx context.Context, input *dto.ListGuidesAdminInput) (*dto.ListGuidesAdminOutput, error) {
	q := dto.ToAdminQueryOptions(input.AdminPaginationQuery)
	result, err := h.guideAdminUC.ListGuides(ctx, q)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	items := make([]dto.AdminGuideCardDTO, 0, len(result.Data))
	for _, guide := range result.Data {
		items = append(items, dto.ToAdminGuideCardDTO(guide))
	}

	return &dto.ListGuidesAdminOutput{Body: dto.ListGuidesAdminResponseBody{
		Guides:     items,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalItems: result.Total,
		TotalPages: result.TotalPages,
	}}, nil
}

func (h *GuideAdminHandler) HandleGetGuideAdmin(ctx context.Context, input *dto.GetGuideAdminInput) (*dto.GetGuideAdminOutput, error) {
	guide, err := h.guideAdminUC.GetGuideDetail(ctx, input.ID, input.Locale)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.GetGuideAdminOutput{Body: dto.GetGuideAdminResponseBody{Guide: dto.ToAdminGuideDetailDTO(guide)}}, nil
}

func (h *GuideAdminHandler) HandleListGuideStepsAdmin(ctx context.Context, input *dto.ListGuideStepsAdminInput) (*dto.ListGuideStepsAdminOutput, error) {
	q := dto.ToAdminQueryOptions(input.AdminPaginationQuery)
	result, err := h.guideAdminUC.ListGuideSteps(ctx, input.ID, q)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	items := make([]dto.AdminGuideStepDTO, 0, len(result.Data))
	for _, step := range result.Data {
		items = append(items, dto.ToAdminGuideStepDTO(step))
	}

	return &dto.ListGuideStepsAdminOutput{Body: dto.ListGuideStepsAdminResponseBody{
		Steps:      items,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalItems: result.Total,
		TotalPages: result.TotalPages,
	}}, nil
}

// --- Guide ---

func (h *GuideAdminHandler) HandleCreateGuide(ctx context.Context, input *dto.CreateGuideInput) (*dto.CreateGuideOutput, error) {
	result, err := h.guideAdminUC.CreateGuide(ctx, dto.ToCreateGuideInput(input.Body))
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.CreateGuideOutput{Body: dto.CreateGuideResponseBody{ID: result.ID}}, nil
}

func (h *GuideAdminHandler) HandleUpdateGuide(ctx context.Context, input *dto.UpdateGuideInput) (*dto.UpdateGuideOutput, error) {
	if err := h.guideAdminUC.UpdateGuide(ctx, input.ID, dto.ToUpdateGuideInput(input.Body)); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UpdateGuideOutput{Body: dto.UpdateGuideResponseBody{Message: "Guide updated"}}, nil
}

func (h *GuideAdminHandler) HandleDeleteGuide(ctx context.Context, input *dto.DeleteGuideInput) (*dto.DeleteGuideOutput, error) {
	if err := h.guideAdminUC.DeleteGuide(ctx, input.ID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.DeleteGuideOutput{Body: dto.DeleteGuideResponseBody{Message: "Guide deleted"}}, nil
}

func (h *GuideAdminHandler) HandleAddGuideCondition(ctx context.Context, input *dto.AddGuideConditionInput) (*dto.AddGuideConditionOutput, error) {
	if err := h.guideAdminUC.AddGuideCondition(ctx, input.GuideID, usecase.ConditionInput{
		ConditionType:  input.Body.ConditionType,
		Operator:       input.Body.Operator,
		ConditionValue: input.Body.ConditionValue,
		IsInverse:      input.Body.IsInverse,
	}); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.AddGuideConditionOutput{Body: dto.AddGuideConditionResponseBody{Message: "Condition added"}}, nil
}

func (h *GuideAdminHandler) HandleRemoveGuideCondition(ctx context.Context, input *dto.RemoveGuideConditionInput) (*dto.RemoveGuideConditionOutput, error) {
	if err := h.guideAdminUC.RemoveGuideCondition(ctx, input.CondID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.RemoveGuideConditionOutput{Body: dto.RemoveGuideConditionResponseBody{Message: "Condition removed"}}, nil
}

func (h *GuideAdminHandler) HandleSetGuideTranslations(ctx context.Context, input *dto.SetGuideTranslationsInput) (*dto.SetGuideTranslationsOutput, error) {
	translations := make([]usecase.TranslationInput, 0, len(input.Body.Translations))
	for _, t := range input.Body.Translations {
		translations = append(translations, usecase.TranslationInput{
			Language:    t.Language,
			Name:        t.Name,
			Description: t.Description,
		})
	}
	if err := h.guideAdminUC.SetGuideTranslations(ctx, input.ID, translations); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.SetGuideTranslationsOutput{Body: dto.SetGuideTranslationsResponseBody{Message: "Translations updated"}}, nil
}

// --- Step ---

func (h *GuideAdminHandler) HandleCreateStep(ctx context.Context, input *dto.CreateStepInput) (*dto.CreateStepOutput, error) {
	result, err := h.guideAdminUC.CreateStep(ctx, dto.ToCreateStepInput(input.Body))
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.CreateStepOutput{Body: dto.CreateStepResponseBody{ID: result.ID}}, nil
}

func (h *GuideAdminHandler) HandleUpdateStep(ctx context.Context, input *dto.UpdateStepInput) (*dto.UpdateStepOutput, error) {
	if err := h.guideAdminUC.UpdateStep(ctx, input.ID, dto.ToUpdateStepInput(input.Body)); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UpdateStepOutput{Body: dto.UpdateStepResponseBody{Message: "Step updated"}}, nil
}

func (h *GuideAdminHandler) HandleDeleteStep(ctx context.Context, input *dto.DeleteStepInput) (*dto.DeleteStepOutput, error) {
	if err := h.guideAdminUC.DeleteStep(ctx, input.ID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.DeleteStepOutput{Body: dto.DeleteStepResponseBody{Message: "Step deleted"}}, nil
}

func (h *GuideAdminHandler) HandleReorderSteps(ctx context.Context, input *dto.ReorderStepsInput) (*dto.ReorderStepsOutput, error) {
	if err := h.guideAdminUC.ReorderSteps(ctx, input.Body.GuideID, input.Body.StepIDs); err != nil {
		if errors.Is(err, guideerror.ErrIncompleteStepList) {
			return nil, apperrors.BadRequestError("guide.errors.incompleteStepList")
		}
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.ReorderStepsOutput{Body: dto.ReorderStepsResponseBody{Message: "Steps reordered"}}, nil
}

func (h *GuideAdminHandler) HandleAddStepCondition(ctx context.Context, input *dto.AddStepConditionInput) (*dto.AddStepConditionOutput, error) {
	if err := h.guideAdminUC.AddStepCondition(ctx, input.StepID, usecase.ConditionInput{
		ConditionType:  input.Body.ConditionType,
		Operator:       input.Body.Operator,
		ConditionValue: input.Body.ConditionValue,
		IsInverse:      input.Body.IsInverse,
	}); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.AddStepConditionOutput{Body: dto.AddStepConditionResponseBody{Message: "Condition added"}}, nil
}

func (h *GuideAdminHandler) HandleRemoveStepCondition(ctx context.Context, input *dto.RemoveStepConditionInput) (*dto.RemoveStepConditionOutput, error) {
	if err := h.guideAdminUC.RemoveStepCondition(ctx, input.CondID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.RemoveStepConditionOutput{Body: dto.RemoveStepConditionResponseBody{Message: "Condition removed"}}, nil
}

func (h *GuideAdminHandler) HandleAddStepDependency(ctx context.Context, input *dto.AddStepDependencyInput) (*dto.AddStepDependencyOutput, error) {
	if err := h.guideAdminUC.AddStepDependency(ctx, input.StepID, input.Body.RequiredStepID, input.Body.DependencyType); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.AddStepDependencyOutput{Body: dto.AddStepDependencyResponseBody{Message: "Dependency added"}}, nil
}

func (h *GuideAdminHandler) HandleRemoveStepDependency(ctx context.Context, input *dto.RemoveStepDependencyInput) (*dto.RemoveStepDependencyOutput, error) {
	if err := h.guideAdminUC.RemoveStepDependency(ctx, input.DepID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.RemoveStepDependencyOutput{Body: dto.RemoveStepDependencyResponseBody{Message: "Dependency removed"}}, nil
}

func (h *GuideAdminHandler) HandleSetStepTranslations(ctx context.Context, input *dto.SetStepTranslationsInput) (*dto.SetStepTranslationsOutput, error) {
	translations := make([]usecase.StepTranslationInput, 0, len(input.Body.Translations))
	for _, t := range input.Body.Translations {
		translations = append(translations, usecase.StepTranslationInput{
			Language:        t.Language,
			Title:           t.Title,
			Description:     t.Description,
			DetailedContent: t.DetailedContent,
		})
	}
	if err := h.guideAdminUC.SetStepTranslations(ctx, input.ID, translations); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.SetStepTranslationsOutput{Body: dto.SetStepTranslationsResponseBody{Message: "Translations updated"}}, nil
}

func (h *GuideAdminHandler) HandleGetStepVersions(ctx context.Context, input *dto.GetStepVersionsInput) (*dto.GetStepVersionsOutput, error) {
	q := dto.ToQueryOptions(input.Page, input.PageSize)
	versions, err := h.guideAdminUC.GetStepVersions(ctx, input.StepID, q)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	result := make([]dto.StepVersionDTO, 0, len(versions))
	for _, v := range versions {
		result = append(result, dto.ToStepVersionDTO(v))
	}
	return &dto.GetStepVersionsOutput{Body: dto.GetStepVersionsResponseBody{Versions: result}}, nil
}

func (h *GuideAdminHandler) HandleRevertStepToVersion(ctx context.Context, input *dto.RevertStepToVersionInput) (*dto.RevertStepToVersionOutput, error) {
	if err := h.guideAdminUC.RevertStepToVersion(ctx, input.StepID, input.Version); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.RevertStepToVersionOutput{Body: dto.RevertStepToVersionResponseBody{Message: "Step reverted to version"}}, nil
}

// --- Journey ---

func (h *GuideAdminHandler) HandleInvalidateUserJourney(ctx context.Context, input *dto.InvalidateUserJourneyInput) (*dto.InvalidateUserJourneyOutput, error) {
	if err := h.journeyAdminUC.InvalidateUserJourney(ctx, input.Body.GuideID, input.UserID, input.Body.GuideID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.InvalidateUserJourneyOutput{Body: dto.InvalidateUserJourneyResponseBody{Message: "User journey invalidated"}}, nil
}

func (h *GuideAdminHandler) HandleInvalidateAllJourneys(ctx context.Context, input *dto.InvalidateAllJourneysInput) (*dto.InvalidateAllJourneysOutput, error) {
	if err := h.journeyAdminUC.InvalidateAllJourneysForGuide(ctx, input.GuideID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.InvalidateAllJourneysOutput{Body: dto.InvalidateAllJourneysResponseBody{Message: "All journeys invalidated"}}, nil
}
