package routes

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/handler"
	"github.com/danielgtaylor/huma/v2"
)

type RouteDependencies struct {
	IngestionHandler        *handler.IngestionHandler
	StatusHandler           *handler.StatusHandler
	AskHandler              *handler.AskHandler
	AuthMiddleware          func(huma.Context, func(huma.Context))
	AccountStatusMiddleware func(huma.Context, func(huma.Context))
}

func RegisterRoutes(api huma.API, deps RouteDependencies) {
	registerIngestionRoutes(api, deps)
	registerStatusRoutes(api, deps)
	registerAskRoutes(api, deps)
}
