package handler

import (
	"context"
	"errors"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/delivery/dto"
	notiferror "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/danielgtaylor/huma/v2"
)

type ScheduledAlertHandler struct {
	scheduledUC usecase.UserScheduledNotificationUsecase
}

func NewScheduledAlertHandler(scheduledUC usecase.UserScheduledNotificationUsecase) *ScheduledAlertHandler {
	return &ScheduledAlertHandler{scheduledUC: scheduledUC}
}

func (h *ScheduledAlertHandler) HandleList(ctx context.Context, input *struct{}) (*dto.ListScheduledAlertsOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	alerts, err := h.scheduledUC.List(ctx, accountID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.ListScheduledAlertsOutput{Body: dto.ListScheduledAlertsResponseBody{
		Data: dto.ToScheduledAlertResponses(alerts),
	}}, nil
}

func (h *ScheduledAlertHandler) HandleCreate(ctx context.Context, input *dto.CreateScheduledAlertInput) (*dto.CreateScheduledAlertOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	notif, err := h.scheduledUC.Schedule(ctx, accountID, dto.ToCreateScheduledAlertInput(input.Body))
	if err != nil {
		if errors.Is(err, notiferror.ErrMaxScheduledAlertsReached) {
			return nil, huma.NewError(403, "Upgrade to Pro to create more than 3 scheduled alerts")
		}
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.CreateScheduledAlertOutput{Body: dto.CreateScheduledAlertResponseBody{
		ID:      notif.ID,
		Message: "Scheduled alert created",
	}}, nil
}

func (h *ScheduledAlertHandler) HandleCancel(ctx context.Context, input *dto.CancelScheduledAlertInput) (*dto.CancelScheduledAlertOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.scheduledUC.Cancel(ctx, accountID, input.ID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.CancelScheduledAlertOutput{Body: dto.CancelScheduledAlertResponseBody{
		Message: "Scheduled alert cancelled",
	}}, nil
}

func (h *ScheduledAlertHandler) HandleReschedule(ctx context.Context, input *dto.RescheduleScheduledAlertInput) (*dto.RescheduleScheduledAlertOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.scheduledUC.Reschedule(ctx, accountID, input.ID, usecase.RescheduleUserNotificationInput{
		ScheduledFor: input.Body.ScheduledFor,
	}); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.RescheduleScheduledAlertOutput{Body: dto.RescheduleScheduledAlertResponseBody{
		Message: "Scheduled alert rescheduled",
	}}, nil
}

func (h *ScheduledAlertHandler) HandleListTemplates(ctx context.Context, input *struct{}) (*dto.ListScheduledTemplatesOutput, error) {
	templates, err := h.scheduledUC.ListTemplates(ctx)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.ListScheduledTemplatesOutput{Body: dto.ListScheduledTemplatesResponseBody{
		Data: dto.ToScheduledTemplateResponses(templates),
	}}, nil
}
