package routes

import (
	"github.com/danielgtaylor/huma/v2"
)

func RegisterLibraryRoutes(api huma.API, deps RouteDependencies) {
	base := "/api/v1/library"

	huma.Register(api, huma.Operation{
		OperationID: "listCategories",
		Method:      "GET",
		Path:        base + "/categories",
		Summary:     "List categories",
		Description: "Lists active categories as a tree, optionally localized.",
		Tags:        []string{"Library"},
	}, deps.ViewHandler.HandleListCategories)

	huma.Register(api, huma.Operation{
		OperationID: "listTemplateGroups",
		Method:      "GET",
		Path:        base + "/templates",
		Summary:     "List template groups",
		Description: "Lists active template groups with optional filters.",
		Tags:        []string{"Library"},
	}, deps.ViewHandler.HandleListTemplateGroups)

	huma.Register(api, huma.Operation{
		OperationID: "getTemplateGroup",
		Method:      "GET",
		Path:        base + "/templates/{slug}",
		Summary:     "Get template group",
		Description: "Gets a template group with its language variants.",
		Tags:        []string{"Library"},
	}, deps.ViewHandler.HandleGetTemplateGroup)

	huma.Register(api, huma.Operation{
		OperationID: "downloadTemplate",
		Method:      "GET",
		Path:        base + "/templates/{slug}/download",
		Summary:     "Download template",
		Description: "Generates a presigned download URL for a template.",
		Tags:        []string{"Library"},
	}, deps.ViewHandler.HandleDownloadTemplate)

	huma.Register(api, huma.Operation{
		OperationID: "listMyDownloads",
		Method:      "GET",
		Path:        base + "/downloads",
		Summary:     "My downloads",
		Description: "Lists the authenticated user's download history.",
		Tags:        []string{"Library"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.ViewHandler.HandleListMyDownloads)
}
