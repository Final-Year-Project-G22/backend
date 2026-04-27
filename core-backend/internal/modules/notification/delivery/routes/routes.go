package routes

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/delivery/handler"
	"github.com/danielgtaylor/huma/v2"
)

type RouteDependencies struct {
	AdminHandler *handler.NotificationAdminHandler
}

func RegisterRoutes(api huma.API, deps RouteDependencies) {
	RegisterAdminNotificationRoutes(api, deps)
}
