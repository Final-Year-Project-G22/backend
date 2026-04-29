package notification

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/google/uuid"
)

type IAMGlobalPreferenceReaderAdapter struct {
	prefRepo repository.NotificationPreferenceRepository
}

func NewIAMGlobalPreferenceReaderAdapter(prefRepo repository.NotificationPreferenceRepository) *IAMGlobalPreferenceReaderAdapter {
	return &IAMGlobalPreferenceReaderAdapter{prefRepo: prefRepo}
}

func (a *IAMGlobalPreferenceReaderAdapter) IsNotificationEnabled(ctx context.Context, accountID uuid.UUID, channel string) (bool, error) {
	pref, err := a.prefRepo.GetByAccountID(ctx, accountID)
	if err != nil || pref == nil {
		return true, nil
	}
	switch channel {
	case "email":
		return pref.EnableEmailNotification, nil
	case "sms":
		return pref.EnableSMSNotification, nil
	case "push":
		return pref.EnablePushNotification, nil
	default:
		return true, nil
	}
}
