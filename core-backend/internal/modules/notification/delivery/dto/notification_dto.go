package dto

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	"github.com/google/uuid"
)

// --- Inbox List ---

type ListInboxInput struct {
	Page     int `query:"page" doc:"Page number"`
	PageSize int `query:"pageSize" doc:"Items per page"`
}

type ListInboxOutput struct {
	Body ListInboxResponseBody
}

type ListInboxResponseBody struct {
	Data       []InboxEntryResponse `json:"data" doc:"Inbox entries"`
	Total      int64                `json:"total" doc:"Total count"`
	Page       int                  `json:"page" doc:"Current page"`
	PageSize   int                  `json:"pageSize" doc:"Items per page"`
	TotalPages int                  `json:"totalPages" doc:"Total pages"`
}

type InboxEntryResponse struct {
	ID           uuid.UUID                   `json:"id" doc:"Inbox entry ID"`
	ActionUrl    *string                     `json:"actionUrl,omitempty" doc:"Action URL"`
	IsRead       bool                        `json:"isRead" doc:"Whether the notification has been read"`
	IsArchived   bool                        `json:"isArchived" doc:"Whether the notification has been archived"`
	ExpiresAt    *time.Time                  `json:"expiresAt,omitempty" doc:"Expiration time"`
	Notification NotificationSummaryResponse `json:"notification" doc:"Notification details"`
}

type NotificationSummaryResponse struct {
	Title            string                  `json:"title" doc:"Notification title"`
	Content          string                  `json:"content" doc:"Notification content"`
	NotificationType entity.NotificationType `json:"notificationType" doc:"Notification type"`
	Channel          entity.Channel          `json:"channel" doc:"Delivery channel"`
	SentAt           time.Time               `json:"sentAt" doc:"Time the notification was sent"`
	DeliveredAt      *time.Time              `json:"deliveredAt,omitempty" doc:"Time the notification was delivered"`
	ReadAt           *time.Time              `json:"readAt,omitempty" doc:"Time the notification was read"`
}

// --- Unread Count ---

type UnreadCountOutput struct {
	Body UnreadCountResponseBody
}

type UnreadCountResponseBody struct {
	Count int64 `json:"count" doc:"Number of unread notifications"`
}

// --- Mark As Read ---

type MarkAsReadInput struct {
	ID uuid.UUID `path:"id" doc:"Inbox entry ID"`
}

type MarkAsReadOutput struct {
	Body MarkAsReadResponseBody
}

type MarkAsReadResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Mark All As Read ---

type MarkAllAsReadOutput struct {
	Body MarkAllAsReadResponseBody
}

type MarkAllAsReadResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Mark Category As Read ---

type MarkCategoryAsReadInput struct {
	Category entity.NotificationCategory `path:"category" doc:"Notification category"`
}

type MarkCategoryAsReadOutput struct {
	Body MarkCategoryAsReadResponseBody
}

type MarkCategoryAsReadResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Archive ---

type ArchiveNotificationInput struct {
	ID uuid.UUID `path:"id" doc:"Inbox entry ID"`
}

type ArchiveNotificationOutput struct {
	Body ArchiveNotificationResponseBody
}

type ArchiveNotificationResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Delete ---

type DeleteNotificationInput struct {
	ID uuid.UUID `path:"id" doc:"Inbox entry ID"`
}

type DeleteNotificationOutput struct {
	Body DeleteNotificationResponseBody
}

type DeleteNotificationResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- History List ---

type ListHistoryInput struct {
	Page     int `query:"page" doc:"Page number"`
	PageSize int `query:"pageSize" doc:"Items per page"`
}

type ListHistoryOutput struct {
	Body ListHistoryResponseBody
}

type ListHistoryResponseBody struct {
	Data       []HistoryEntryResponse `json:"data" doc:"History entries"`
	Total      int64                  `json:"total" doc:"Total count"`
	Page       int                    `json:"page" doc:"Current page"`
	PageSize   int                    `json:"pageSize" doc:"Items per page"`
	TotalPages int                    `json:"totalPages" doc:"Total pages"`
}

type HistoryEntryResponse struct {
	ID               uuid.UUID               `json:"id" doc:"History entry ID"`
	NotificationType entity.NotificationType `json:"notificationType" doc:"Notification type"`
	Channel          entity.Channel          `json:"channel" doc:"Delivery channel"`
	Title            string                  `json:"title" doc:"Notification title"`
	Content          string                  `json:"content" doc:"Notification content"`
	ActionUrl        *string                 `json:"actionUrl,omitempty" doc:"Action URL"`
	SentAt           time.Time               `json:"sentAt" doc:"Time the notification was sent"`
	DeliveredAt      *time.Time              `json:"deliveredAt,omitempty" doc:"Time the notification was delivered"`
	ReadAt           *time.Time              `json:"readAt,omitempty" doc:"Time the notification was read"`
	ClickedAt        *time.Time              `json:"clickedAt,omitempty" doc:"Time the notification was clicked"`
	DeliveryStatus   entity.DeliveryStatus   `json:"deliveryStatus" doc:"Current delivery status"`
	FailureReason    *string                 `json:"failureReason,omitempty" doc:"Failure reason if delivery failed"`
}

// --- History Detail ---

type GetHistoryInput struct {
	ID uuid.UUID `path:"id" doc:"History entry ID"`
}

type GetHistoryOutput struct {
	Body HistoryEntryResponse
}

// --- Preferences ---

type SetPreferenceRequest struct {
	NotificationType entity.NotificationType `json:"notificationType" doc:"Notification type"`
	Channel          entity.Channel          `json:"channel" doc:"Channel"`
	IsEnabled        bool                    `json:"isEnabled" doc:"Whether the notification type is enabled for this channel"`
	QuietHoursStart  *time.Time              `json:"quietHoursStart,omitempty" doc:"Quiet hours start time"`
	QuietHoursEnd    *time.Time              `json:"quietHoursEnd,omitempty" doc:"Quiet hours end time"`
}

type SetPreferenceInput struct {
	Body SetPreferenceRequest
}

type SetPreferenceOutput struct {
	Body SetPreferenceResponseBody
}

type SetPreferenceResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type ListPreferencesOutput struct {
	Body []PreferenceResponse
}

type PreferenceResponse struct {
	NotificationType entity.NotificationType `json:"notificationType" doc:"Notification type"`
	Channel          entity.Channel          `json:"channel" doc:"Channel"`
	IsEnabled        bool                    `json:"isEnabled" doc:"Whether enabled"`
	QuietHoursStart  *time.Time              `json:"quietHoursStart,omitempty" doc:"Quiet hours start time"`
	QuietHoursEnd    *time.Time              `json:"quietHoursEnd,omitempty" doc:"Quiet hours end time"`
}

type DeletePreferenceInput struct {
	NotificationType entity.NotificationType `path:"type" doc:"Notification type"`
	Channel          entity.Channel          `path:"channel" doc:"Channel"`
}

type DeletePreferenceOutput struct {
	Body DeletePreferenceResponseBody
}

type DeletePreferenceResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Mutes ---

type MuteAccountRequest struct {
	MutedAccountID uuid.UUID  `json:"mutedAccountId" doc:"Account ID to mute"`
	MuteUntil      *time.Time `json:"muteUntil,omitempty" doc:"Optional expiration time"`
	Reason         *string    `json:"reason,omitempty" doc:"Reason for muting"`
}

type MuteAccountInput struct {
	Body MuteAccountRequest
}

type MuteAccountOutput struct {
	Body MuteAccountResponseBody
}

type MuteAccountResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type ListMutesInput struct {
	Page     int `query:"page" doc:"Page number"`
	PageSize int `query:"pageSize" doc:"Items per page"`
}

type ListMutesOutput struct {
	Body ListMutesResponseBody
}

type ListMutesResponseBody struct {
	Data       []MuteEntryResponse `json:"data" doc:"Muted accounts"`
	Total      int64               `json:"total" doc:"Total count"`
	Page       int                 `json:"page" doc:"Current page"`
	PageSize   int                 `json:"pageSize" doc:"Items per page"`
	TotalPages int                 `json:"totalPages" doc:"Total pages"`
}

type MuteEntryResponse struct {
	ID             uuid.UUID  `json:"id" doc:"Mute entry ID"`
	MutedAccountID uuid.UUID  `json:"mutedAccountId" doc:"Muted account ID"`
	MuteUntil      *time.Time `json:"muteUntil,omitempty" doc:"Expiration time"`
	Reason         *string    `json:"reason,omitempty" doc:"Reason for muting"`
	CreatedAt      *time.Time `json:"createdAt" doc:"When the mute was created"`
}

type UnmuteAccountInput struct {
	MutedAccountID uuid.UUID `path:"accountId" doc:"Muted account ID"`
}

type UnmuteAccountOutput struct {
	Body UnmuteAccountResponseBody
}

type UnmuteAccountResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Devices ---

type RegisterDeviceRequest struct {
	DeviceType  entity.DeviceType `json:"deviceType" doc:"Device type"`
	DeviceToken string            `json:"deviceToken" doc:"Device token"`
	PushToken   *string           `json:"pushToken,omitempty" doc:"Push notification token"`
	DeviceName  *string           `json:"deviceName,omitempty" doc:"Device name"`
	DeviceModel *string           `json:"deviceModel,omitempty" doc:"Device model"`
	OSVersion   *string           `json:"osVersion,omitempty" doc:"OS version"`
	AppVersion  *string           `json:"appVersion,omitempty" doc:"App version"`
}

type RegisterDeviceInput struct {
	Body RegisterDeviceRequest
}

type RegisterDeviceOutput struct {
	Body RegisterDeviceResponseBody
}

type RegisterDeviceResponseBody struct {
	ID      uuid.UUID `json:"id" doc:"Device ID"`
	Message string    `json:"message" doc:"Success message"`
}

type ListDevicesOutput struct {
	Body []DeviceResponse
}

type DeviceResponse struct {
	ID           uuid.UUID         `json:"id" doc:"Device ID"`
	DeviceType   entity.DeviceType `json:"deviceType" doc:"Device type"`
	DeviceToken  string            `json:"deviceToken" doc:"Device token"`
	PushToken    *string           `json:"pushToken,omitempty" doc:"Push notification token"`
	DeviceName   *string           `json:"deviceName,omitempty" doc:"Device name"`
	DeviceModel  *string           `json:"deviceModel,omitempty" doc:"Device model"`
	OSVersion    *string           `json:"osVersion,omitempty" doc:"OS version"`
	AppVersion   *string           `json:"appVersion,omitempty" doc:"App version"`
	IsActive     bool              `json:"isActive" doc:"Whether the device is active"`
	LastActiveAt *time.Time        `json:"lastActiveAt,omitempty" doc:"Last active time"`
}

type UpdateDeviceRequest struct {
	PushToken  *string `json:"pushToken,omitempty" doc:"Push notification token"`
	DeviceName *string `json:"deviceName,omitempty" doc:"Device name"`
	OSVersion  *string `json:"osVersion,omitempty" doc:"OS version"`
	AppVersion *string `json:"appVersion,omitempty" doc:"App version"`
	IsActive   *bool   `json:"isActive,omitempty" doc:"Whether the device is active"`
}

type UpdateDeviceInput struct {
	ID   uuid.UUID `path:"id" doc:"Device ID"`
	Body UpdateDeviceRequest
}

type UpdateDeviceOutput struct {
	Body DeviceResponse
}

type DeactivateDeviceInput struct {
	ID uuid.UUID `path:"id" doc:"Device ID"`
}

type DeactivateDeviceOutput struct {
	Body DeactivateDeviceResponseBody
}

type DeactivateDeviceResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Mappers ---

func ToSetPreferenceInput(body SetPreferenceRequest) usecase.SetPreferenceInput {
	return usecase.SetPreferenceInput{
		NotificationType: body.NotificationType,
		Channel:          body.Channel,
		IsEnabled:        body.IsEnabled,
		QuietHoursStart:  body.QuietHoursStart,
		QuietHoursEnd:    body.QuietHoursEnd,
	}
}

func ToMuteAccountInput(body MuteAccountRequest) usecase.MuteAccountInput {
	return usecase.MuteAccountInput{
		MutedAccountID: body.MutedAccountID,
		MuteUntil:      body.MuteUntil,
		Reason:         body.Reason,
	}
}

func ToRegisterDeviceInput(body RegisterDeviceRequest) usecase.RegisterDeviceInput {
	return usecase.RegisterDeviceInput{
		DeviceType:  body.DeviceType,
		DeviceToken: body.DeviceToken,
		PushToken:   body.PushToken,
		DeviceName:  body.DeviceName,
		DeviceModel: body.DeviceModel,
		OSVersion:   body.OSVersion,
		AppVersion:  body.AppVersion,
	}
}

func ToUpdateDeviceInput(body UpdateDeviceRequest) usecase.UpdateDeviceInput {
	return usecase.UpdateDeviceInput{
		PushToken:  body.PushToken,
		DeviceName: body.DeviceName,
		OSVersion:  body.OSVersion,
		AppVersion: body.AppVersion,
		IsActive:   body.IsActive,
	}
}

func ToPreferenceResponse(pref *entity.UserNotificationPreference) PreferenceResponse {
	return PreferenceResponse{
		NotificationType: pref.NotificationType,
		Channel:          pref.Channel,
		IsEnabled:        pref.IsEnabled,
		QuietHoursStart:  pref.QuietHoursStart,
		QuietHoursEnd:    pref.QuietHoursEnd,
	}
}

func ToPreferenceResponses(prefs []*entity.UserNotificationPreference) []PreferenceResponse {
	if len(prefs) == 0 {
		return nil
	}
	resp := make([]PreferenceResponse, 0, len(prefs))
	for _, p := range prefs {
		resp = append(resp, ToPreferenceResponse(p))
	}
	return resp
}

func ToMuteEntryResponse(mute *entity.MutedAccount) MuteEntryResponse {
	return MuteEntryResponse{
		ID:             mute.ID,
		MutedAccountID: mute.MutedAccountID,
		MuteUntil:      mute.MuteUntil,
		Reason:         mute.Reason,
		CreatedAt:      mute.CreatedAt,
	}
}

func ToMuteEntryResponses(mutes []*entity.MutedAccount) []MuteEntryResponse {
	if len(mutes) == 0 {
		return nil
	}
	resp := make([]MuteEntryResponse, 0, len(mutes))
	for _, m := range mutes {
		resp = append(resp, ToMuteEntryResponse(m))
	}
	return resp
}

func ToDeviceResponse(device *entity.UserDevice) DeviceResponse {
	return DeviceResponse{
		ID:           device.ID,
		DeviceType:   device.DeviceType,
		DeviceToken:  device.DeviceToken,
		PushToken:    device.PushToken,
		DeviceName:   device.DeviceName,
		DeviceModel:  device.DeviceModel,
		OSVersion:    device.OSVersion,
		AppVersion:   device.AppVersion,
		IsActive:     device.IsActive,
		LastActiveAt: device.LastActiveAt,
	}
}

func ToDeviceResponses(devices []*entity.UserDevice) []DeviceResponse {
	if len(devices) == 0 {
		return nil
	}
	resp := make([]DeviceResponse, 0, len(devices))
	for _, d := range devices {
		resp = append(resp, ToDeviceResponse(d))
	}
	return resp
}

// --- Existing mappers below ---

func ToInboxEntryResponse(inbox *entity.UserNotificationInbox) InboxEntryResponse {
	resp := InboxEntryResponse{
		ID: inbox.ID,

		ActionUrl:  inbox.ActionUrl,
		IsRead:     inbox.IsRead,
		IsArchived: inbox.IsArchived,
		ExpiresAt:  inbox.ExpiresAt,
	}
	if inbox.NotificationHistory.ID != uuid.Nil {
		resp.Notification = NotificationSummaryResponse{
			Title:            inbox.NotificationHistory.Title,
			Content:          inbox.NotificationHistory.Content,
			NotificationType: inbox.NotificationHistory.NotificationType,
			Channel:          inbox.NotificationHistory.Channel,
			SentAt:           inbox.NotificationHistory.SentAt,
			DeliveredAt:      inbox.NotificationHistory.DeliveredAt,
			ReadAt:           inbox.NotificationHistory.ReadAt,
		}
	}
	return resp
}

func ToInboxEntryResponses(inboxes []*entity.UserNotificationInbox) []InboxEntryResponse {
	if len(inboxes) == 0 {
		return nil
	}
	resp := make([]InboxEntryResponse, 0, len(inboxes))
	for _, inbox := range inboxes {
		resp = append(resp, ToInboxEntryResponse(inbox))
	}
	return resp
}

func ToHistoryEntryResponse(history *entity.NotificationHistory) HistoryEntryResponse {
	return HistoryEntryResponse{
		ID:               history.ID,
		NotificationType: history.NotificationType,
		Channel:          history.Channel,
		Title:            history.Title,
		Content:          history.Content,
		ActionUrl:        history.ActionUrl,
		SentAt:           history.SentAt,
		DeliveredAt:      history.DeliveredAt,
		ReadAt:           history.ReadAt,
		ClickedAt:        history.ClickedAt,
		DeliveryStatus:   history.DeliveryStatus,
		FailureReason:    history.FailureReason,
	}
}

func ToHistoryEntryResponses(histories []*entity.NotificationHistory) []HistoryEntryResponse {
	if len(histories) == 0 {
		return nil
	}
	resp := make([]HistoryEntryResponse, 0, len(histories))
	for _, h := range histories {
		resp = append(resp, ToHistoryEntryResponse(h))
	}
	return resp
}
