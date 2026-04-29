package routes

import (
	"github.com/gin-gonic/gin"
)

const webhookBase = "/api/v1/webhooks"

func RegisterWebhookRoutes(engine *gin.Engine, deps RouteDependencies) {
	engine.POST(webhookBase+"/resend", deps.WebhookHandler.HandleResendWebhook)
}
