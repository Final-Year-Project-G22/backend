package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/google/uuid"
)

type PreferenceUsecase interface {
	GetAccountPreferences(ctx context.Context, accountID uuid.UUID) (*AccountPreferencesView, error)
	UpdateAccountPreference(ctx context.Context, accountID uuid.UUID, input UpdateAccountPreferenceInput) (*entity.AccountPreference, error)
	UpdateNotificationPreference(ctx context.Context, accountID uuid.UUID, input UpdateNotificationPreferenceInput) (*entity.NotificationPreference, error)
	UpdateCommunityPreference(ctx context.Context, accountID uuid.UUID, input UpdateCommunityPreferenceInput) (*entity.CommunityPreference, error)
	UpdateAIPreference(ctx context.Context, accountID uuid.UUID, input UpdateAIPreferenceInput) (*entity.AIPreference, error)
	UpdateTemplatePreference(ctx context.Context, accountID uuid.UUID, input UpdateTemplatePreferenceInput) (*entity.TemplatePreference, error)
}

type UpdateAccountPreferenceInput struct {
	Language *string
	Timezone *string
}

type UpdateNotificationPreferenceInput struct {
	EnableEmailNotification *bool
	EnableSMSNotification   *bool
	EnablePushNotification  *bool
	CampaignDigestEnabled   *bool
}

type UpdateCommunityPreferenceInput struct {
	AllowMentions *bool
	AllowReplies  *bool
	DigestEnabled *bool
}

type UpdateAIPreferenceInput struct {
	DefaultModel       *string
	ResponseStyle      *string
	Temperature        *float64
	AllowDataRetention *bool
}

type UpdateTemplatePreferenceInput struct {
	DefaultTemplate *string
	EditorMode      *string
}

type AccountPreferencesView struct {
	AccountPreference      *entity.AccountPreference
	NotificationPreference *entity.NotificationPreference
	CommunityPreference    *entity.CommunityPreference
	AIPreference           *entity.AIPreference
	TemplatePreference     *entity.TemplatePreference
}
