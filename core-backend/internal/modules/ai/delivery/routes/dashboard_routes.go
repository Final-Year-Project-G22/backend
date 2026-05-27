package routes

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/handler"
	"github.com/danielgtaylor/huma/v2"
)

const adminDashboardBase = "/api/v1/admin/dashboard"

func RegisterAIDashboardRoutes(api huma.API, deps RouteDependencies, aiDashboardHandler *handler.AIDashboardHandler) {
	huma.Register(api, huma.Operation{
		OperationID: "getDocumentStats",
		Method:      "GET",
		Path:        adminDashboardBase + "/document-stats",
		Summary:     "Get document ingestion statistics",
		Tags:        []string{"Admin - Dashboard"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, aiDashboardHandler.HandleDocumentStats)

	huma.Register(api, huma.Operation{
		OperationID: "getDocumentStages",
		Method:      "GET",
		Path:        adminDashboardBase + "/document-stages",
		Summary:     "Get document pipeline stage breakdown",
		Tags:        []string{"Admin - Dashboard"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, aiDashboardHandler.HandleDocumentStages)
}
