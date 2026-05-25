package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/dto"
	iamentity "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	iamrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
)

type NotificationPreferenceHandler struct {
	repo iamrepo.NotificationPreferenceRepository
}

func NewNotificationPreferenceHandler(repo iamrepo.NotificationPreferenceRepository) *NotificationPreferenceHandler {
	return &NotificationPreferenceHandler{repo: repo}
}

func (h *NotificationPreferenceHandler) HandleGetNotificationPreferences(ctx context.Context, input *struct{}) (*dto.GetNotificationPreferencesOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	pref, err := h.repo.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	if pref == nil {
		return &dto.GetNotificationPreferencesOutput{
			Body: dto.NotificationPreferencesResponse{
				EmailEnabled: true,
				PushEnabled:  true,
			},
		}, nil
	}
	return &dto.GetNotificationPreferencesOutput{
		Body: dto.NotificationPreferencesResponse{
			EmailEnabled: pref.EnableEmailNotification,
			PushEnabled:  pref.EnablePushNotification,
		},
	}, nil
}

func (h *NotificationPreferenceHandler) HandleUpdateNotificationPreferences(ctx context.Context, input *dto.UpdateNotificationPreferencesInput) (*dto.UpdateNotificationPreferencesOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))

	pref, err := h.repo.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	if pref == nil {
		pref = &iamentity.NotificationPreference{
			AccountID:               accountID,
			EnableEmailNotification: true,
			EnablePushNotification:  false,
		}
	}

	if input.Body.EmailEnabled != nil {
		pref.EnableEmailNotification = *input.Body.EmailEnabled
	}
	if input.Body.PushEnabled != nil {
		pref.EnablePushNotification = *input.Body.PushEnabled
	}

	if pref.ID == uuid.Nil {
		if err := h.repo.Create(ctx, pref); err != nil {
			return nil, apperrors.ToHumaError(ctx, err)
		}
	} else {
		if err := h.repo.Update(ctx, pref); err != nil {
			return nil, apperrors.ToHumaError(ctx, err)
		}
	}

	return &dto.UpdateNotificationPreferencesOutput{
		Body: dto.NotificationPreferencesResponse{
			EmailEnabled: pref.EnableEmailNotification,
			PushEnabled:  pref.EnablePushNotification,
		},
	}, nil
}
