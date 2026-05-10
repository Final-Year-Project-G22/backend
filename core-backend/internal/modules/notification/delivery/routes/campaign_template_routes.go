package routes

import (
	"github.com/danielgtaylor/huma/v2"
)

const campaignTemplateBase = adminNotifBase + "/campaign-templates"

func RegisterCampaignTemplateRoutes(api huma.API, deps RouteDependencies) {
	// --- Campaign Templates ---
	huma.Register(api, huma.Operation{
		OperationID: "createCampaignTemplate",
		Method:      "POST",
		Path:        campaignTemplateBase,
		Summary:     "Create campaign template",
		Description: "Creates a new campaign template.",
		Tags:        []string{"Admin - Campaign Templates"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.CampaignTemplateHandler.HandleCreateCampaignTemplate)

	huma.Register(api, huma.Operation{
		OperationID: "listCampaignTemplates",
		Method:      "GET",
		Path:        campaignTemplateBase,
		Summary:     "List campaign templates",
		Description: "Lists all campaign templates.",
		Tags:        []string{"Admin - Campaign Templates"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.CampaignTemplateHandler.HandleListCampaignTemplates)

	huma.Register(api, huma.Operation{
		OperationID: "getCampaignTemplate",
		Method:      "GET",
		Path:        campaignTemplateBase + "/{id}",
		Summary:     "Get campaign template",
		Description: "Gets a campaign template with its translations.",
		Tags:        []string{"Admin - Campaign Templates"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.CampaignTemplateHandler.HandleGetCampaignTemplate)

	huma.Register(api, huma.Operation{
		OperationID: "updateCampaignTemplate",
		Method:      "PATCH",
		Path:        campaignTemplateBase + "/{id}",
		Summary:     "Update campaign template",
		Description: "Updates a campaign template.",
		Tags:        []string{"Admin - Campaign Templates"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.CampaignTemplateHandler.HandleUpdateCampaignTemplate)

	huma.Register(api, huma.Operation{
		OperationID: "deleteCampaignTemplate",
		Method:      "DELETE",
		Path:        campaignTemplateBase + "/{id}",
		Summary:     "Delete campaign template",
		Description: "Deletes a campaign template.",
		Tags:        []string{"Admin - Campaign Templates"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.CampaignTemplateHandler.HandleDeleteCampaignTemplate)

	// --- Campaign Template Translations ---
	huma.Register(api, huma.Operation{
		OperationID: "addCampaignTemplateTranslation",
		Method:      "POST",
		Path:        campaignTemplateBase + "/{id}/translations",
		Summary:     "Add campaign template translation",
		Description: "Adds a translation to a campaign template.",
		Tags:        []string{"Admin - Campaign Templates"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.CampaignTemplateHandler.HandleAddCampaignTemplateTranslation)

	huma.Register(api, huma.Operation{
		OperationID: "updateCampaignTemplateTranslation",
		Method:      "PATCH",
		Path:        campaignTemplateBase + "/{id}/translations/{lang}",
		Summary:     "Update campaign template translation",
		Description: "Updates an existing translation for a campaign template.",
		Tags:        []string{"Admin - Campaign Templates"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.CampaignTemplateHandler.HandleUpdateCampaignTemplateTranslation)

	huma.Register(api, huma.Operation{
		OperationID: "deleteCampaignTemplateTranslation",
		Method:      "DELETE",
		Path:        campaignTemplateBase + "/{id}/translations/{lang}",
		Summary:     "Delete campaign template translation",
		Description: "Deletes a translation from a campaign template.",
		Tags:        []string{"Admin - Campaign Templates"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.CampaignTemplateHandler.HandleDeleteCampaignTemplateTranslation)
}
