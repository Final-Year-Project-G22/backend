package notification

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
)

type EntityProvider struct{}

func NewEntityProvider() *EntityProvider {
	return &EntityProvider{}
}

func (e *EntityProvider) Entities() []any {
	return []any{
		&entity.NotificationTemplate{},
		&entity.NotificationTemplateTranslation{},
		&entity.UserNotificationPreference{},
		&entity.MutedAccount{},
		&entity.UserDevice{},
		&entity.NotificationQueue{},
		&entity.NotificationHistory{},
		&entity.UserNotificationInbox{},
		&entity.NotificationCampaign{},
		&entity.EmailDeliveryLog{},
		&entity.NotificationOutbox{},
	}
}

func (e *EntityProvider) ModuleName() string {
	return "notification"
}
