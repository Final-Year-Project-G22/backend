package routes

import (
	"github.com/danielgtaylor/huma/v2"
)

func RegisterLibraryAdminRoutes(api huma.API, deps RouteDependencies) {
	adminBase := "/api/v1/admin/library"

	// --- Categories ---
	huma.Register(api, huma.Operation{
		OperationID: "LibraryListAllCategories",
		Method:      "GET",
		Path:        adminBase + "/categories",
		Summary:     "List all categories",
		Description: "Lists categories, optionally including inactive.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
	}, deps.AdminHandler.HandleListAllCategories)

	huma.Register(api, huma.Operation{
		OperationID: "LibraryCreateCategory",
		Method:      "POST",
		Path:        adminBase + "/categories",
		Summary:     "Create category",
		Description: "Creates a new library category.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.WritePermissionMiddleware},
	}, deps.AdminHandler.HandleCreateCategory)

	huma.Register(api, huma.Operation{
		OperationID: "LibraryGetCategory",
		Method:      "GET",
		Path:        adminBase + "/categories/{id}",
		Summary:     "Get category",
		Description: "Gets a category with translations.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
	}, deps.AdminHandler.HandleGetCategory)

	huma.Register(api, huma.Operation{
		OperationID: "LibraryUpdateCategory",
		Method:      "PATCH",
		Path:        adminBase + "/categories/{id}",
		Summary:     "Update category",
		Description: "Updates a category's metadata.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.UpdatePermissionMiddleware},
	}, deps.AdminHandler.HandleUpdateCategory)

	huma.Register(api, huma.Operation{
		OperationID: "LibraryDeleteCategory",
		Method:      "DELETE",
		Path:        adminBase + "/categories/{id}",
		Summary:     "Delete category",
		Description: "Soft-deletes a category. Fails if it has active template groups.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.DeletePermissionMiddleware},
	}, deps.AdminHandler.HandleDeleteCategory)

	// --- Category Translations ---
	huma.Register(api, huma.Operation{
		OperationID: "LibraryAddCategoryTranslation",
		Method:      "POST",
		Path:        adminBase + "/categories/{id}/translations",
		Summary:     "Add category translation",
		Description: "Adds a translation for a category.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.WritePermissionMiddleware},
	}, deps.AdminHandler.HandleAddCategoryTranslation)

	huma.Register(api, huma.Operation{
		OperationID: "LibraryUpdateCategoryTranslation",
		Method:      "PATCH",
		Path:        adminBase + "/categories/{id}/translations/{lang}",
		Summary:     "Update category translation",
		Description: "Updates an existing category translation.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.UpdatePermissionMiddleware},
	}, deps.AdminHandler.HandleUpdateCategoryTranslation)

	huma.Register(api, huma.Operation{
		OperationID: "LibraryDeleteCategoryTranslation",
		Method:      "DELETE",
		Path:        adminBase + "/categories/{id}/translations/{lang}",
		Summary:     "Delete category translation",
		Description: "Removes a category translation.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.DeletePermissionMiddleware},
	}, deps.AdminHandler.HandleDeleteCategoryTranslation)

	// --- Template Groups ---
	huma.Register(api, huma.Operation{
		OperationID: "LibraryListAllTemplateGroups",
		Method:      "GET",
		Path:        adminBase + "/template-groups",
		Summary:     "List template groups",
		Description: "Lists all template groups with optional category filter.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
	}, deps.AdminHandler.HandleListAllTemplateGroups)

	huma.Register(api, huma.Operation{
		OperationID: "LibraryCreateTemplateGroup",
		Method:      "POST",
		Path:        adminBase + "/template-groups",
		Summary:     "Create template group",
		Description: "Creates a new template group with metadata.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.WritePermissionMiddleware},
	}, deps.AdminHandler.HandleCreateTemplateGroup)

	huma.Register(api, huma.Operation{
		OperationID: "LibraryAdminGetTemplateGroup",
		Method:      "GET",
		Path:        adminBase + "/template-groups/{groupId}",
		Summary:     "Get template group",
		Description: "Gets a template group with all templates.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
	}, deps.AdminHandler.HandleGetTemplateGroup)

	huma.Register(api, huma.Operation{
		OperationID: "LibraryUpdateTemplateGroup",
		Method:      "PATCH",
		Path:        adminBase + "/template-groups/{groupId}",
		Summary:     "Update template group",
		Description: "Updates a template group's metadata.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.UpdatePermissionMiddleware},
	}, deps.AdminHandler.HandleUpdateTemplateGroup)

	huma.Register(api, huma.Operation{
		OperationID: "LibraryDeleteTemplateGroup",
		Method:      "DELETE",
		Path:        adminBase + "/template-groups/{groupId}",
		Summary:     "Delete template group",
		Description: "Soft-deletes a template group.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.DeletePermissionMiddleware},
	}, deps.AdminHandler.HandleDeleteTemplateGroup)

	// --- Templates ---
	huma.Register(api, huma.Operation{
		OperationID: "LibraryListTemplatesByGroup",
		Method:      "GET",
		Path:        adminBase + "/template-groups/{groupId}/templates",
		Summary:     "List templates by group",
		Description: "Lists all template variants for a group.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
	}, deps.AdminHandler.HandleListTemplatesByGroup)

	huma.Register(api, huma.Operation{
		OperationID: "LibraryCreateTemplate",
		Method:      "POST",
		Path:        adminBase + "/template-groups/{groupId}/templates",
		Summary:     "Create template",
		Description: "Uploads a file and creates a template language variant.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.WritePermissionMiddleware},
	}, deps.AdminHandler.HandleCreateTemplate)

	huma.Register(api, huma.Operation{
		OperationID: "LibraryGetTemplate",
		Method:      "GET",
		Path:        adminBase + "/templates/{templateId}",
		Summary:     "Get template",
		Description: "Gets a template by ID.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
	}, deps.AdminHandler.HandleGetTemplate)

	huma.Register(api, huma.Operation{
		OperationID: "LibraryUpdateTemplate",
		Method:      "PATCH",
		Path:        adminBase + "/templates/{templateId}",
		Summary:     "Update template",
		Description: "Updates template metadata or replaces the file.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.UpdatePermissionMiddleware},
	}, deps.AdminHandler.HandleUpdateTemplate)

	huma.Register(api, huma.Operation{
		OperationID: "LibraryDeleteTemplate",
		Method:      "DELETE",
		Path:        adminBase + "/templates/{templateId}",
		Summary:     "Delete template",
		Description: "Soft-deletes a template.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.DeletePermissionMiddleware},
	}, deps.AdminHandler.HandleDeleteTemplate)

	// --- Interactive Forms ---
	huma.Register(api, huma.Operation{
		OperationID: "LibraryGetInteractiveForm",
		Method:      "GET",
		Path:        adminBase + "/templates/{templateId}/interactive-form",
		Summary:     "Get interactive form",
		Description: "Gets the interactive form for a template.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
	}, deps.AdminHandler.HandleGetInteractiveForm)

	huma.Register(api, huma.Operation{
		OperationID: "LibraryCreateInteractiveForm",
		Method:      "POST",
		Path:        adminBase + "/templates/{templateId}/interactive-form",
		Summary:     "Create interactive form",
		Description: "Creates an interactive form for an interactive template.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.WritePermissionMiddleware},
	}, deps.AdminHandler.HandleCreateInteractiveForm)

	huma.Register(api, huma.Operation{
		OperationID: "LibraryUpdateInteractiveForm",
		Method:      "PATCH",
		Path:        adminBase + "/interactive-forms/{id}",
		Summary:     "Update interactive form",
		Description: "Updates an interactive form's layout.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.UpdatePermissionMiddleware},
	}, deps.AdminHandler.HandleUpdateInteractiveForm)

	huma.Register(api, huma.Operation{
		OperationID: "LibraryDeleteInteractiveForm",
		Method:      "DELETE",
		Path:        adminBase + "/interactive-forms/{id}",
		Summary:     "Delete interactive form",
		Description: "Soft-deletes an interactive form.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.DeletePermissionMiddleware},
	}, deps.AdminHandler.HandleDeleteInteractiveForm)

	// --- Download Logs ---
	huma.Register(api, huma.Operation{
		OperationID: "LibraryGetDownloadLogs",
		Method:      "GET",
		Path:        adminBase + "/downloads",
		Summary:     "Get download logs",
		Description: "Lists download history with optional group filter.",
		Tags:        []string{"Admin - Library"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
	}, deps.AdminHandler.HandleGetDownloadLogs)
}
