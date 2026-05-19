package routes

import (
	"github.com/danielgtaylor/huma/v2"
)

func RegisterScheduledAlertRoutes(api huma.API, deps RouteDependencies) {
	huma.Register(api, huma.Operation{
		OperationID: "listScheduledAlerts",
		Method:      "GET",
		Path:        notifBase + "/scheduled",
		Summary:     "List scheduled alerts",
		Description: "Lists all scheduled alerts for the authenticated user.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.ScheduledAlertHandler.HandleList)

	huma.Register(api, huma.Operation{
		OperationID: "createScheduledAlert",
		Method:      "POST",
		Path:        notifBase + "/scheduled",
		Summary:     "Create scheduled alert",
		Description: "Creates a new scheduled alert for the authenticated user.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.ScheduledAlertHandler.HandleCreate)

	huma.Register(api, huma.Operation{
		OperationID: "cancelScheduledAlert",
		Method:      "PATCH",
		Path:        notifBase + "/scheduled/{id}/cancel",
		Summary:     "Cancel scheduled alert",
		Description: "Cancels a pending scheduled alert.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.ScheduledAlertHandler.HandleCancel)

	huma.Register(api, huma.Operation{
		OperationID: "rescheduleScheduledAlert",
		Method:      "PATCH",
		Path:        notifBase + "/scheduled/{id}/reschedule",
		Summary:     "Reschedule scheduled alert",
		Description: "Reschedules a pending scheduled alert to a new date.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.ScheduledAlertHandler.HandleReschedule)

	huma.Register(api, huma.Operation{
		OperationID: "listScheduledAlertTemplates",
		Method:      "GET",
		Path:        notifBase + "/scheduled/templates",
		Summary:     "List scheduled alert templates",
		Description: "Lists all available templates for creating scheduled alerts.",
		Tags:        []string{"Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.ScheduledAlertHandler.HandleListTemplates)
}
