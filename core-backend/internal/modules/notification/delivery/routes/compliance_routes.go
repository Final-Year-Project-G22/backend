package routes

import (
	"github.com/danielgtaylor/huma/v2"
)

const complianceBase = "/api/v1/compliance"

func RegisterComplianceRoutes(api huma.API, deps RouteDependencies) {
	huma.Register(api, huma.Operation{
		OperationID: "listComplianceEntries",
		Method:      "GET",
		Path:        complianceBase + "/entries",
		Summary:     "List compliance entries",
		Description: "Lists all compliance entries for a business profile.",
		Tags:        []string{"Compliance"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.ComplianceHandler.HandleList)

	huma.Register(api, huma.Operation{
		OperationID: "createComplianceEntry",
		Method:      "POST",
		Path:        complianceBase + "/entries",
		Summary:     "Create compliance entry",
		Description: "Creates a new compliance entry for a business profile.",
		Tags:        []string{"Compliance"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.ComplianceHandler.HandleCreate)

	huma.Register(api, huma.Operation{
		OperationID: "updateComplianceEntry",
		Method:      "PATCH",
		Path:        complianceBase + "/entries/{id}",
		Summary:     "Update compliance entry",
		Description: "Updates a compliance entry.",
		Tags:        []string{"Compliance"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.ComplianceHandler.HandleUpdate)

	huma.Register(api, huma.Operation{
		OperationID: "deleteComplianceEntry",
		Method:      "DELETE",
		Path:        complianceBase + "/entries/{id}",
		Summary:     "Delete compliance entry",
		Description: "Deletes a compliance entry.",
		Tags:        []string{"Compliance"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.ComplianceHandler.HandleDelete)

	huma.Register(api, huma.Operation{
		OperationID: "getComplianceCalendar",
		Method:      "GET",
		Path:        complianceBase + "/calendar",
		Summary:     "Get compliance calendar",
		Description: "Returns upcoming compliance deadlines and scheduled alerts for the authenticated user.",
		Tags:        []string{"Compliance"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.ComplianceHandler.HandleGetCalendar)
}
