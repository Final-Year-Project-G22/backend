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
}
