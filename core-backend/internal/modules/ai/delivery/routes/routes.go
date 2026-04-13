package routes

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/handler"
	"github.com/danielgtaylor/huma/v2"
)

type RouteDependencies struct {
	IngestionHandler *handler.IngestionHandler
}

func RegisterRoutes(api huma.API, deps RouteDependencies) {
	registerIngestionRoutes(api, deps)
}
