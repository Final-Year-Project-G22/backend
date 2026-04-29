package routes

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/delivery/handler"
	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"
)

type RouteDependencies struct {
	AdminHandler            *handler.NotificationAdminHandler
	NotificationHandler     *handler.NotificationHandler
	WebhookHandler          *handler.WebhookHandler
	AuthMiddleware          func(huma.Context, func(huma.Context))
	AccountStatusMiddleware func(huma.Context, func(huma.Context))
}

func RegisterRoutes(api huma.API, engine *gin.Engine, deps RouteDependencies) {
	RegisterAdminNotificationRoutes(api, deps)
	RegisterNotificationRoutes(api, deps)
	RegisterWebhookRoutes(engine, deps)
}
