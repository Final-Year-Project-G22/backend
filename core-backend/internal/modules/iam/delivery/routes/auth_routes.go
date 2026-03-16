package routes

import (
	"github.com/danielgtaylor/huma/v2"
)

const (
	apiV1Base = "/api/v1"
	authBase  = apiV1Base + "/auth"
)

func registerAuthRoutes(api huma.API, deps RouteDependencies) {
	huma.Register(api, huma.Operation{
		OperationID: "register",
		Method:      "POST",
		Path:        authBase + "/register",
		Summary:     "Register a new user",
		Description: "Creates a new user account and returns authentication tokens. The user is automatically logged in after registration.",
		Tags:        []string{"Authentication"},
	}, deps.AuthHandler.HandleRegister)

	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      "POST",
		Path:        authBase + "/login",
		Summary:     "Log in a user",
		Description: "Authenticates a user with email and password, returns authentication tokens.",
		Tags:        []string{"Authentication"},
	}, deps.AuthHandler.HandleLogin)

	huma.Register(api, huma.Operation{
		OperationID: "refresh",
		Method:      "POST",
		Path:        authBase + "/refresh",
		Summary:     "Refresh access token",
		Description: "Uses the refresh token cookie to issue new access and refresh tokens. Implements token rotation for security.",
		Tags:        []string{"Authentication"},
	}, deps.AuthHandler.HandleRefresh)

	huma.Register(api, huma.Operation{
		OperationID:   "logout",
		Method:        "POST",
		Path:          authBase + "/logout",
		Summary:       "Log out current session",
		Description:   "Revokes the current session and clears the refresh token cookie.",
		Tags:          []string{"Authentication"},
		Middlewares:   huma.Middlewares{deps.AuthMiddleware},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: 200,
	}, deps.AuthHandler.HandleLogout)

	huma.Register(api, huma.Operation{
		OperationID:   "logoutAll",
		Method:        "POST",
		Path:          authBase + "/logout/all",
		Summary:       "Log out all sessions",
		Description:   "Revokes all sessions for the current user's account and clears the refresh token cookie.",
		Tags:          []string{"Authentication"},
		Middlewares:   huma.Middlewares{deps.AuthMiddleware},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: 200,
	}, deps.AuthHandler.HandleLogoutAll)

	registerSecurityScheme(api)
}

func registerSecurityScheme(api huma.API) {
	spec := api.OpenAPI()
	if spec.Components == nil {
		spec.Components = &huma.Components{}
	}
	if spec.Components.SecuritySchemes == nil {
		spec.Components.SecuritySchemes = make(map[string]*huma.SecurityScheme)
	}

	spec.Components.SecuritySchemes["bearerAuth"] = &huma.SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  "JWT access token obtained from login or register endpoints",
	}
}
