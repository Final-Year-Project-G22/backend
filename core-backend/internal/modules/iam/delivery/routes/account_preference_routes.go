package routes

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/handler"
	"github.com/danielgtaylor/huma/v2"
)

const accountPrefBase = "/api/v1/auth/preferences"

type AccountPreferenceRouteDependencies struct {
	Handler                 *handler.AccountPreferenceHandler
	AuthMiddleware          func(huma.Context, func(huma.Context))
	AccountStatusMiddleware func(huma.Context, func(huma.Context))
}

func RegisterAccountPreferenceRoutes(api huma.API, deps AccountPreferenceRouteDependencies) {
	huma.Register(api, huma.Operation{
		OperationID: "getAccountPreferences",
		Method:      "GET",
		Path:        accountPrefBase,
		Summary:     "Get account preferences",
		Description: "Returns the language and timezone preferences for the authenticated user.",
		Tags:        []string{"Authentication"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.Handler.HandleGetAccountPreferences)

	huma.Register(api, huma.Operation{
		OperationID: "updateAccountPreferences",
		Method:      "PUT",
		Path:        accountPrefBase,
		Summary:     "Update account preferences",
		Description: "Updates the language and timezone preferences for the authenticated user.",
		Tags:        []string{"Authentication"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
	}, deps.Handler.HandleUpdateAccountPreferences)
}
