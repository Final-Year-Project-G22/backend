package routes

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/handler"
	"github.com/danielgtaylor/huma/v2"
)

const notifPrefBase = "/api/v1/notifications"

type NotificationPreferenceRouteDependencies struct {
	Handler                 *handler.NotificationPreferenceHandler
	AuthMiddleware          func(huma.Context, func(huma.Context))
	AccountStatusMiddleware func(huma.Context, func(huma.Context))
}

func RegisterNotificationPreferenceRoutes(api huma.API, deps NotificationPreferenceRouteDependencies) {
	huma.Register(api, huma.Operation{
		OperationID: "getNotificationPreferences",
		Method:      "GET",
		Path:        notifPrefBase + "/global-preferences",
		Summary:     "Get notification preferences",
		Description: "Returns the global email and push notification preferences for the authenticated user.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.Handler.HandleGetNotificationPreferences)

	huma.Register(api, huma.Operation{
		OperationID: "updateNotificationPreferences",
		Method:      "PUT",
		Path:        notifPrefBase + "/global-preferences",
		Summary:     "Update notification preferences",
		Description: "Updates the global email and push notification preferences for the authenticated user.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.Handler.HandleUpdateNotificationPreferences)
}
