package routes

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/delivery/handler"
	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"
)

type RouteDependencies struct {
	AdminHandler               *handler.NotificationAdminHandler
	CampaignTemplateHandler    *handler.CampaignTemplateHandler
	NotificationHandler        *handler.NotificationHandler
	WebhookHandler             *handler.WebhookHandler
	SSEHandler                 *handler.SSEHandler
	InboxSSEHandler            *handler.InboxSSEHandler
	ScheduledAlertHandler      *handler.ScheduledAlertHandler
	ComplianceHandler          *handler.ComplianceHandler
	AuthMiddleware             func(huma.Context, func(huma.Context))
	AccountStatusMiddleware    func(huma.Context, func(huma.Context))
	ReadPermissionMiddleware   func(huma.Context, func(huma.Context))
	WritePermissionMiddleware  func(huma.Context, func(huma.Context))
	UpdatePermissionMiddleware func(huma.Context, func(huma.Context))
	DeletePermissionMiddleware func(huma.Context, func(huma.Context))
}

func RegisterRoutes(api huma.API, engine *gin.Engine, deps RouteDependencies) {
	RegisterAdminNotificationRoutes(api, deps)
	RegisterCampaignTemplateRoutes(api, deps)
	RegisterNotificationRoutes(api, deps)
	RegisterScheduledAlertRoutes(api, deps)
	RegisterComplianceRoutes(api, deps)
	RegisterWebhookRoutes(engine, deps)
	RegisterSSERoutes(engine, deps)

	engine.OPTIONS(notifBase+"/inbox/events", deps.InboxSSEHandler.HandleInboxEvents)
	engine.GET(notifBase+"/inbox/events", deps.InboxSSEHandler.HandleInboxEvents)
}
