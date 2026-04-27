package routes

import (
	"github.com/danielgtaylor/huma/v2"
)

const adminNotifBase = "/api/v1/admin/notifications"

func RegisterAdminNotificationRoutes(api huma.API, deps RouteDependencies) {
	huma.Register(api, huma.Operation{
		OperationID: "createTemplate",
		Method:      "POST",
		Path:        adminNotifBase + "/templates",
		Summary:     "Create template",
		Description: "Creates a new notification template.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.AdminHandler.HandleCreateTemplate)

	huma.Register(api, huma.Operation{
		OperationID: "listTemplates",
		Method:      "GET",
		Path:        adminNotifBase + "/templates",
		Summary:     "List templates",
		Description: "Lists notification templates with optional category filter.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.AdminHandler.HandleListTemplates)

	huma.Register(api, huma.Operation{
		OperationID: "getTemplate",
		Method:      "GET",
		Path:        adminNotifBase + "/templates/{id}",
		Summary:     "Get template",
		Description: "Gets a notification template with its translations.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.AdminHandler.HandleGetTemplate)

	huma.Register(api, huma.Operation{
		OperationID: "updateTemplate",
		Method:      "PATCH",
		Path:        adminNotifBase + "/templates/{id}",
		Summary:     "Update template",
		Description: "Updates a notification template. System-managed templates have restricted fields.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.AdminHandler.HandleUpdateTemplate)

	huma.Register(api, huma.Operation{
		OperationID: "deleteTemplate",
		Method:      "DELETE",
		Path:        adminNotifBase + "/templates/{id}",
		Summary:     "Delete template",
		Description: "Soft-deletes a notification template. System-managed templates cannot be deleted.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.AdminHandler.HandleDeleteTemplate)

	huma.Register(api, huma.Operation{
		OperationID: "addTranslation",
		Method:      "POST",
		Path:        adminNotifBase + "/templates/{id}/translations",
		Summary:     "Add translation",
		Description: "Adds a translation to a notification template.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.AdminHandler.HandleAddTranslation)

	huma.Register(api, huma.Operation{
		OperationID: "updateTranslation",
		Method:      "PATCH",
		Path:        adminNotifBase + "/templates/{id}/translations/{lang}",
		Summary:     "Update translation",
		Description: "Updates an existing translation for a notification template.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.AdminHandler.HandleUpdateTranslation)

	huma.Register(api, huma.Operation{
		OperationID: "deleteTranslation",
		Method:      "DELETE",
		Path:        adminNotifBase + "/templates/{id}/translations/{lang}",
		Summary:     "Delete translation",
		Description: "Deletes a translation from a notification template.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.AdminHandler.HandleDeleteTranslation)
}
