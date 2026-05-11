package routes

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/handler"
	"github.com/danielgtaylor/huma/v2"
)

const businessProfileBase = "/api/v1/users/business-profile"

type BusinessProfileRouteDependencies struct {
	BusinessProfileHandler  *handler.BusinessProfileHandler
	AuthMiddleware          func(huma.Context, func(huma.Context))
	AccountStatusMiddleware func(huma.Context, func(huma.Context))
}

func RegisterBusinessProfileRoutes(api huma.API, deps BusinessProfileRouteDependencies) {
	huma.Register(api, huma.Operation{
		OperationID: "getBusinessProfile",
		Method:      "GET",
		Path:        businessProfileBase,
		Summary:     "Get business profile",
		Description: "Retrieves the authenticated user's business profile.",
		Tags:        []string{"Business Profile"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.BusinessProfileHandler.HandleGetBusinessProfile)

	huma.Register(api, huma.Operation{
		OperationID: "createBusinessProfile",
		Method:      "POST",
		Path:        businessProfileBase,
		Summary:     "Create business profile",
		Description: "Creates a business profile for the authenticated user. Accepts slugs for sector and tags, which are resolved server-side. Auto-fills company name, email, and phone from account data if omitted.",
		Tags:        []string{"Business Profile"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.BusinessProfileHandler.HandleCreateBusinessProfile)

	huma.Register(api, huma.Operation{
		OperationID: "updateBusinessProfile",
		Method:      "PUT",
		Path:        businessProfileBase,
		Summary:     "Update business profile",
		Description: "Updates the authenticated user's business profile. Accepts slugs for sector and tags, which are resolved server-side.",
		Tags:        []string{"Business Profile"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.BusinessProfileHandler.HandleUpdateBusinessProfile)
}
