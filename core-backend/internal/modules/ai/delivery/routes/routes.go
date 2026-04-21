package routes

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/handler"
	"github.com/danielgtaylor/huma/v2"
)

type RouteDependencies struct {
	IngestionHandler        *handler.IngestionHandler
	StatusHandler           *handler.StatusHandler
	AskHandler              *handler.AskHandler
	DLQHandler              *handler.DLQHandler
	SSEHandler              *handler.SSEHandler
	ToggleHandler           *handler.ToggleHandler
	AskEnabled              bool
	AuthMiddleware          func(huma.Context, func(huma.Context))
	AccountStatusMiddleware func(huma.Context, func(huma.Context))
}

func RegisterRoutes(api huma.API, deps RouteDependencies) {
	registerIngestionRoutes(api, deps)
	registerStatusRoutes(api, deps)
	if deps.AskEnabled {
		registerAskRoutes(api, deps)
	}
	registerDLQRoutes(api, deps)
	registerIngestionStatusStreamRoute(api, deps)
	registerIngestionStatusDocumentStreamRoute(api, deps)
	registerIngestToggleRoute(api, deps)
}
