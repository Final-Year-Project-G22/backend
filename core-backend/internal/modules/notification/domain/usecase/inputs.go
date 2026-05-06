package usecase

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/google/uuid"
)

type CreateTemplateInput struct {
	Name             string                      `json:"name"`
	Description      *string                     `json:"description,omitempty"`
	NotificationType entity.NotificationType     `json:"notificationType"`
	TemplateGroup    string                      `json:"templateGroup"`
	Priority         entity.NotificationPriority `json:"priority"`
	DefaultContent   map[string]interface{}      `json:"defaultContent"`
	VariablesSchema  *map[string]interface{}     `json:"variablesSchema,omitempty"`
	DefaultTTL       *int                        `json:"defaultTtl,omitempty"`
}

type UpdateTemplateInput struct {
	Name            *string                      `json:"name,omitempty"`
	Description     *string                      `json:"description,omitempty"`
	Priority        *entity.NotificationPriority `json:"priority,omitempty"`
	DefaultContent  *map[string]interface{}      `json:"defaultContent,omitempty"`
	VariablesSchema *map[string]interface{}      `json:"variablesSchema,omitempty"`
	DefaultTTL      *int                         `json:"defaultTtl,omitempty"`
}

type CreateTemplateTranslationInput struct {
	TemplateID uuid.UUID              `json:"templateId"`
	Language   string                 `json:"language"`
	Subject    string                 `json:"subject"`
	Content    map[string]interface{} `json:"content"`
}

type UpdateTemplateTranslationInput struct {
	Subject *string                 `json:"subject,omitempty"`
	Content *map[string]interface{} `json:"content,omitempty"`
}

type SetPreferenceInput struct {
	NotificationType entity.NotificationType `json:"notificationType"`
	Channel          entity.Channel          `json:"channel"`
	IsEnabled        bool                    `json:"isEnabled"`
	QuietHoursStart  *time.Time              `json:"quietHoursStart,omitempty"`
	QuietHoursEnd    *time.Time              `json:"quietHoursEnd,omitempty"`
}

type MuteAccountInput struct {
	MutedAccountID uuid.UUID  `json:"mutedAccountId"`
	MuteUntil      *time.Time `json:"muteUntil,omitempty"`
	Reason         *string    `json:"reason,omitempty"`
}

type RegisterDeviceInput struct {
	DeviceType  entity.DeviceType `json:"deviceType"`
	DeviceToken string            `json:"deviceToken"`
	PushToken   *string           `json:"pushToken,omitempty"`
	DeviceName  *string           `json:"deviceName,omitempty"`
	DeviceModel *string           `json:"deviceModel,omitempty"`
	OSVersion   *string           `json:"osVersion,omitempty"`
	AppVersion  *string           `json:"appVersion,omitempty"`
}

type UpdateDeviceInput struct {
	PushToken  *string `json:"pushToken,omitempty"`
	DeviceName *string `json:"deviceName,omitempty"`
	OSVersion  *string `json:"osVersion,omitempty"`
	AppVersion *string `json:"appVersion,omitempty"`
	IsActive   *bool   `json:"isActive,omitempty"`
}

type ProcessEventInput struct {
	SourceModule     string                  `json:"sourceModule"`
	SourceEvent      string                  `json:"sourceEvent"`
	NotificationType entity.NotificationType `json:"notificationType"`
	ChannelPolicy    string                  `json:"channelPolicy"`
	Channel          *entity.Channel         `json:"channel,omitempty"`
	AccountID        uuid.UUID               `json:"accountId"`
	Variables        map[string]string       `json:"variables"`
	Metadata         map[string]interface{}  `json:"metadata"`
	ExpiresAt        *time.Time              `json:"expiresAt,omitempty"`
}

type SendNotificationInput struct {
	NotificationType entity.NotificationType `json:"notificationType"`
	AccountID        uuid.UUID               `json:"accountId"`
	Channel          entity.Channel          `json:"channel"`
	Variables        map[string]string       `json:"variables"`
	Metadata         map[string]interface{}  `json:"metadata"`
	ScheduledFor     *time.Time              `json:"scheduledFor,omitempty"`
	ExpiresAt        *time.Time              `json:"expiresAt,omitempty"`
}

type CreateCampaignInput struct {
	Name               string                  `json:"name"`
	Description        *string                 `json:"description,omitempty"`
	CampaignType       entity.CampaignType     `json:"campaignType"`
	TargetSegment      *map[string]interface{} `json:"targetSegment,omitempty"`
	CampaignTemplateID uuid.UUID               `json:"campaignTemplateId"`
	ScheduledFor       *time.Time              `json:"scheduledFor,omitempty"`
	SectorIDs          []uuid.UUID             `json:"sectorIds,omitempty"`
	TagIDs             []uuid.UUID             `json:"tagIds,omitempty"`
	Region             *string                 `json:"region,omitempty"`
	Stage              *string                 `json:"stage,omitempty"`
}

type UpdateCampaignInput struct {
	Name          *string                 `json:"name,omitempty"`
	Description   *string                 `json:"description,omitempty"`
	TargetSegment *map[string]interface{} `json:"targetSegment,omitempty"`
	ScheduledFor  *time.Time              `json:"scheduledFor,omitempty"`
	SectorIDs     []uuid.UUID             `json:"sectorIds,omitempty"`
	TagIDs        []uuid.UUID             `json:"tagIds,omitempty"`
	Region        *string                 `json:"region,omitempty"`
	Stage         *string                 `json:"stage,omitempty"`
}

type ScheduleCampaignInput struct {
	CampaignID uuid.UUID `json:"campaignId"`
}

type CampaignDetail struct {
	Campaign         *entity.NotificationCampaign
	CampaignTemplate *entity.CampaignTemplate
	CreatedByName    string
	CreatedByEmail   string
}

type ResendWebhookEvent struct {
	EventType      string    `json:"eventType"`
	EmailID        string    `json:"emailId"`
	RecipientEmail string    `json:"recipientEmail"`
	OccurredAt     time.Time `json:"occurredAt"`
	BounceReason   *string   `json:"bounceReason,omitempty"`
}
