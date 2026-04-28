package routes

import (
	"github.com/danielgtaylor/huma/v2"
)

const notifBase = "/api/v1/notifications"

func RegisterNotificationRoutes(api huma.API, deps RouteDependencies) {
	// --- Inbox ---
	huma.Register(api, huma.Operation{
		OperationID: "listInbox",
		Method:      "GET",
		Path:        notifBase + "/inbox",
		Summary:     "List inbox",
		Description: "Lists the authenticated user's inbox with optional category filter.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.NotificationHandler.HandleListInbox)

	huma.Register(api, huma.Operation{
		OperationID: "getUnreadCount",
		Method:      "GET",
		Path:        notifBase + "/inbox/unread-count",
		Summary:     "Get unread count",
		Description: "Returns the number of unread notifications for the authenticated user.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.NotificationHandler.HandleGetUnreadCount)

	huma.Register(api, huma.Operation{
		OperationID: "markAsRead",
		Method:      "PATCH",
		Path:        notifBase + "/inbox/{id}/read",
		Summary:     "Mark as read",
		Description: "Marks a single inbox notification as read.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.NotificationHandler.HandleMarkAsRead)

	huma.Register(api, huma.Operation{
		OperationID: "markAllAsRead",
		Method:      "POST",
		Path:        notifBase + "/inbox/read-all",
		Summary:     "Mark all as read",
		Description: "Marks all inbox notifications as read for the authenticated user.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.NotificationHandler.HandleMarkAllAsRead)

	huma.Register(api, huma.Operation{
		OperationID: "markCategoryAsRead",
		Method:      "POST",
		Path:        notifBase + "/inbox/category/{category}/read",
		Summary:     "Mark category as read",
		Description: "Marks all inbox notifications in a category as read.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.NotificationHandler.HandleMarkCategoryAsRead)

	huma.Register(api, huma.Operation{
		OperationID: "archiveNotification",
		Method:      "PATCH",
		Path:        notifBase + "/inbox/{id}/archive",
		Summary:     "Archive notification",
		Description: "Archives a single inbox notification.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.NotificationHandler.HandleArchiveNotification)

	huma.Register(api, huma.Operation{
		OperationID: "deleteNotification",
		Method:      "DELETE",
		Path:        notifBase + "/inbox/{id}",
		Summary:     "Delete notification",
		Description: "Soft-deletes a single inbox notification.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.NotificationHandler.HandleDeleteNotification)

	// --- History ---
	huma.Register(api, huma.Operation{
		OperationID: "listHistory",
		Method:      "GET",
		Path:        notifBase + "/history",
		Summary:     "List history",
		Description: "Lists the authenticated user's notification history.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.NotificationHandler.HandleListHistory)

	huma.Register(api, huma.Operation{
		OperationID: "getHistoryDetail",
		Method:      "GET",
		Path:        notifBase + "/history/{id}",
		Summary:     "Get history detail",
		Description: "Gets a single notification history entry.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.NotificationHandler.HandleGetHistoryDetail)

	// --- Preferences ---
	huma.Register(api, huma.Operation{
		OperationID: "listPreferences",
		Method:      "GET",
		Path:        notifBase + "/preferences",
		Summary:     "List preferences",
		Description: "Lists all notification preference overrides for the authenticated user.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.NotificationHandler.HandleListPreferences)

	huma.Register(api, huma.Operation{
		OperationID: "setPreference",
		Method:      "PUT",
		Path:        notifBase + "/preferences",
		Summary:     "Set preference",
		Description: "Sets a notification preference override for a type+channel combination.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.NotificationHandler.HandleSetPreference)

	huma.Register(api, huma.Operation{
		OperationID: "deletePreference",
		Method:      "DELETE",
		Path:        notifBase + "/preferences/{type}/{channel}",
		Summary:     "Delete preference",
		Description: "Deletes a notification preference override, reverting to defaults.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.NotificationHandler.HandleDeletePreference)

	// --- Mutes ---
	huma.Register(api, huma.Operation{
		OperationID: "listMutes",
		Method:      "GET",
		Path:        notifBase + "/mutes",
		Summary:     "List mutes",
		Description: "Lists all muted accounts for the authenticated user.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.NotificationHandler.HandleListMutes)

	huma.Register(api, huma.Operation{
		OperationID: "muteAccount",
		Method:      "POST",
		Path:        notifBase + "/mutes",
		Summary:     "Mute account",
		Description: "Mutes an account so their notifications are not shown in the inbox.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.NotificationHandler.HandleMuteAccount)

	huma.Register(api, huma.Operation{
		OperationID: "unmuteAccount",
		Method:      "DELETE",
		Path:        notifBase + "/mutes/{accountId}",
		Summary:     "Unmute account",
		Description: "Unmutes a previously muted account.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.NotificationHandler.HandleUnmuteAccount)

	// --- Devices ---
	huma.Register(api, huma.Operation{
		OperationID: "listDevices",
		Method:      "GET",
		Path:        notifBase + "/devices",
		Summary:     "List devices",
		Description: "Lists all registered devices for the authenticated user.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.NotificationHandler.HandleListDevices)

	huma.Register(api, huma.Operation{
		OperationID: "registerDevice",
		Method:      "POST",
		Path:        notifBase + "/devices",
		Summary:     "Register device",
		Description: "Registers a new device or updates an existing one by device token.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.NotificationHandler.HandleRegisterDevice)

	huma.Register(api, huma.Operation{
		OperationID: "updateDevice",
		Method:      "PATCH",
		Path:        notifBase + "/devices/{id}",
		Summary:     "Update device",
		Description: "Updates device metadata such as push token or device name.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.NotificationHandler.HandleUpdateDevice)

	huma.Register(api, huma.Operation{
		OperationID: "deactivateDevice",
		Method:      "DELETE",
		Path:        notifBase + "/devices/{id}",
		Summary:     "Deactivate device",
		Description: "Deactivates a registered device.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.NotificationHandler.HandleDeactivateDevice)
}
