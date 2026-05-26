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
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.WritePermissionMiddleware},
	}, deps.AdminHandler.HandleCreateTemplate)

	huma.Register(api, huma.Operation{
		OperationID: "listTemplates",
		Method:      "GET",
		Path:        adminNotifBase + "/templates",
		Summary:     "List templates",
		Description: "Lists notification templates with optional category filter.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
	}, deps.AdminHandler.HandleListTemplates)

	huma.Register(api, huma.Operation{
		OperationID: "getTemplate",
		Method:      "GET",
		Path:        adminNotifBase + "/templates/{id}",
		Summary:     "Get template",
		Description: "Gets a notification template with its translations.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
	}, deps.AdminHandler.HandleGetTemplate)

	huma.Register(api, huma.Operation{
		OperationID: "updateTemplate",
		Method:      "PATCH",
		Path:        adminNotifBase + "/templates/{id}",
		Summary:     "Update template",
		Description: "Updates a notification template. System-managed templates have restricted fields.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.UpdatePermissionMiddleware},
	}, deps.AdminHandler.HandleUpdateTemplate)

	huma.Register(api, huma.Operation{
		OperationID: "deleteTemplate",
		Method:      "DELETE",
		Path:        adminNotifBase + "/templates/{id}",
		Summary:     "Delete template",
		Description: "Soft-deletes a notification template. System-managed templates cannot be deleted.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.DeletePermissionMiddleware},
	}, deps.AdminHandler.HandleDeleteTemplate)

	huma.Register(api, huma.Operation{
		OperationID: "addTranslation",
		Method:      "POST",
		Path:        adminNotifBase + "/templates/{id}/translations",
		Summary:     "Add translation",
		Description: "Adds a translation to a notification template.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.WritePermissionMiddleware},
	}, deps.AdminHandler.HandleAddTranslation)

	huma.Register(api, huma.Operation{
		OperationID: "updateTranslation",
		Method:      "PATCH",
		Path:        adminNotifBase + "/templates/{id}/translations/{lang}",
		Summary:     "Update translation",
		Description: "Updates an existing translation for a notification template.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.UpdatePermissionMiddleware},
	}, deps.AdminHandler.HandleUpdateTranslation)

	huma.Register(api, huma.Operation{
		OperationID: "deleteTranslation",
		Method:      "DELETE",
		Path:        adminNotifBase + "/templates/{id}/translations/{lang}",
		Summary:     "Delete translation",
		Description: "Deletes a translation from a notification template.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.DeletePermissionMiddleware},
	}, deps.AdminHandler.HandleDeleteTranslation)

	// --- Monitoring ---
	huma.Register(api, huma.Operation{
		OperationID: "getQueueStatus",
		Method:      "GET",
		Path:        adminNotifBase + "/queue/status",
		Summary:     "Get queue status",
		Description: "Returns notification queue counts by status.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
	}, deps.AdminHandler.HandleGetQueueStatus)

	huma.Register(api, huma.Operation{
		OperationID: "retryFailed",
		Method:      "POST",
		Path:        adminNotifBase + "/queue/retry",
		Summary:     "Retry failed",
		Description: "Retries failed notification queue items.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.WritePermissionMiddleware},
	}, deps.AdminHandler.HandleRetryFailed)

	// --- Campaigns ---
	huma.Register(api, huma.Operation{
		OperationID: "createCampaign",
		Method:      "POST",
		Path:        adminNotifBase + "/campaigns",
		Summary:     "Create campaign",
		Description: "Creates a new notification campaign in draft status.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.WritePermissionMiddleware},
	}, deps.AdminHandler.HandleCreateCampaign)

	huma.Register(api, huma.Operation{
		OperationID: "listCampaigns",
		Method:      "GET",
		Path:        adminNotifBase + "/campaigns",
		Summary:     "List campaigns",
		Description: "Lists notification campaigns with optional status filter.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
	}, deps.AdminHandler.HandleListCampaigns)

	huma.Register(api, huma.Operation{
		OperationID: "getCampaign",
		Method:      "GET",
		Path:        adminNotifBase + "/campaigns/{id}",
		Summary:     "Get campaign",
		Description: "Gets a notification campaign by ID.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
	}, deps.AdminHandler.HandleGetCampaign)

	huma.Register(api, huma.Operation{
		OperationID: "updateCampaign",
		Method:      "PATCH",
		Path:        adminNotifBase + "/campaigns/{id}",
		Summary:     "Update campaign",
		Description: "Updates a draft notification campaign.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.UpdatePermissionMiddleware},
	}, deps.AdminHandler.HandleUpdateCampaign)

	huma.Register(api, huma.Operation{
		OperationID: "scheduleCampaign",
		Method:      "POST",
		Path:        adminNotifBase + "/campaigns/{id}/schedule",
		Summary:     "Schedule campaign",
		Description: "Schedules a draft campaign, resolving segment filters to a static recipient list.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.WritePermissionMiddleware},
	}, deps.AdminHandler.HandleScheduleCampaign)

	huma.Register(api, huma.Operation{
		OperationID: "cancelCampaign",
		Method:      "POST",
		Path:        adminNotifBase + "/campaigns/{id}/cancel",
		Summary:     "Cancel campaign",
		Description: "Cancels a scheduled or sending campaign.",
		Tags:        []string{"Admin - Notifications"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.WritePermissionMiddleware},
	}, deps.AdminHandler.HandleCancelCampaign)
}
