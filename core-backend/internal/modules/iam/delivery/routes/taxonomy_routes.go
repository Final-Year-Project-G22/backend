package routes

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/handler"
	"github.com/danielgtaylor/huma/v2"
)

const taxonomyBase = "/api/v1/taxonomy"

func RegisterTaxonomyRoutes(api huma.API, deps TaxonomyRouteDependencies) {
	huma.Register(api, huma.Operation{
		OperationID: "listTaxonomySectors",
		Method:      "GET",
		Path:        taxonomyBase + "/sectors",
		Summary:     "List sectors",
		Description: "Lists all available sectors.",
		Tags:        []string{"Taxonomy"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.TaxonomyHandler.HandleListSectors)

	huma.Register(api, huma.Operation{
		OperationID: "getTaxonomySector",
		Method:      "GET",
		Path:        taxonomyBase + "/sectors/{id}",
		Summary:     "Get sector",
		Description: "Retrieves a single sector by ID.",
		Tags:        []string{"Taxonomy"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.TaxonomyHandler.HandleGetSector)

	huma.Register(api, huma.Operation{
		OperationID: "listTaxonomyTags",
		Method:      "GET",
		Path:        taxonomyBase + "/tags",
		Summary:     "List tags",
		Description: "Lists all available tags.",
		Tags:        []string{"Taxonomy"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.TaxonomyHandler.HandleListTags)

	huma.Register(api, huma.Operation{
		OperationID: "getTaxonomyTag",
		Method:      "GET",
		Path:        taxonomyBase + "/tags/{id}",
		Summary:     "Get tag",
		Description: "Retrieves a single tag by ID.",
		Tags:        []string{"Taxonomy"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.TaxonomyHandler.HandleGetTag)
}

type TaxonomyRouteDependencies struct {
	TaxonomyHandler         *handler.TaxonomyHandler
	AuthMiddleware          func(huma.Context, func(huma.Context))
	AccountStatusMiddleware func(huma.Context, func(huma.Context))
}
