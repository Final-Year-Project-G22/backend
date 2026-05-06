package routes

import (
	"github.com/gin-gonic/gin"
)

const sseCampaignBase = adminNotifBase + "/campaigns/events"

func RegisterSSERoutes(engine *gin.Engine, deps RouteDependencies) {
	engine.GET(sseCampaignBase, deps.SSEHandler.HandleCampaignEvents)
}
