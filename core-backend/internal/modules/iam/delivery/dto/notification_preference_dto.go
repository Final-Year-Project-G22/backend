package dto

// --- Notification Preferences ---

type NotificationPreferencesResponse struct {
	EmailEnabled bool `json:"emailEnabled"`
	PushEnabled  bool `json:"pushEnabled"`
}

type UpdateNotificationPreferencesRequest struct {
	EmailEnabled *bool `json:"emailEnabled,omitempty"`
	PushEnabled  *bool `json:"pushEnabled,omitempty"`
}

type UpdateNotificationPreferencesInput struct {
	Body UpdateNotificationPreferencesRequest
}

type UpdateNotificationPreferencesOutput struct {
	Body NotificationPreferencesResponse
}

type GetNotificationPreferencesOutput struct {
	Body NotificationPreferencesResponse
}
