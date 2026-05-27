package routes

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/handler"
	"github.com/danielgtaylor/huma/v2"
)

const dashboardBase = adminBase + "/dashboard"

func RegisterDashboardRoutes(api huma.API, deps AdminRouteDependencies, dashboardHandler *handler.DashboardHandler) {
	huma.Register(api, huma.Operation{
		OperationID: "getUserStats",
		Method:      "GET",
		Path:        dashboardBase + "/user-stats",
		Summary:     "Get user statistics",
		Tags:        []string{"Admin - Dashboard"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, dashboardHandler.HandleUserStats)

	huma.Register(api, huma.Operation{
		OperationID: "getSessionStats",
		Method:      "GET",
		Path:        dashboardBase + "/session-stats",
		Summary:     "Get session statistics",
		Tags:        []string{"Admin - Dashboard"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, dashboardHandler.HandleSessionStats)

	huma.Register(api, huma.Operation{
		OperationID: "getUserGrowth",
		Method:      "GET",
		Path:        dashboardBase + "/user-growth",
		Summary:     "Get user growth by tier",
		Tags:        []string{"Admin - Dashboard"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, dashboardHandler.HandleUserGrowth)

	huma.Register(api, huma.Operation{
		OperationID: "getSystemOverview",
		Method:      "GET",
		Path:        dashboardBase + "/system-overview",
		Summary:     "Get system overview indicators",
		Tags:        []string{"Admin - Dashboard"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, dashboardHandler.HandleSystemOverview)

	huma.Register(api, huma.Operation{
		OperationID: "getActivityLogs",
		Method:      "GET",
		Path:        dashboardBase + "/activity-logs",
		Summary:     "Get recent admin activity logs",
		Tags:        []string{"Admin - Dashboard"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, dashboardHandler.HandleActivityLogs)

	huma.Register(api, huma.Operation{
		OperationID: "getReportStats",
		Method:      "GET",
		Path:        dashboardBase + "/report-stats",
		Summary:     "Get content report statistics",
		Tags:        []string{"Admin - Dashboard"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, dashboardHandler.HandleReportStats)
}
