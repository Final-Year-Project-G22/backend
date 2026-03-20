package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

const (
	apiV1Base = "/api/v1"
	authBase  = apiV1Base + "/auth"
)

func RegisterAuthRoutes(api huma.API, deps RouteDependencies) {
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

	registerOAuthRoutes(api, deps)

	huma.Register(api, huma.Operation{
		OperationID: "AccountPassword",
		Method:      "PUT",
		Path:        authBase + "/user/updatePassword",
		Summary:     "Update account password",
		Description: "Updates account password",
		Tags:        []string{"Authentication"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.AuthHandler.HandleAccountPasswordUpdate)

	registerSecurityScheme(api)
}

func registerOAuthRoutes(api huma.API, deps RouteDependencies) {
	huma.Register(api, huma.Operation{
		OperationID: "getOAuthProviders",
		Method:      "GET",
		Path:        authBase + "/oauth/providers",
		Summary:     "List OAuth providers",
		Description: "Returns a list of available OAuth providers for authentication.",
		Tags:        []string{"OAuth"},
	}, deps.OAuthHandler.HandleGetProviders)

	huma.Register(api, huma.Operation{
		OperationID:   "initiateOAuthLogin",
		Method:        "GET",
		Path:          authBase + "/oauth/login/{provider}",
		Summary:       "Initiate OAuth login",
		Description:   "Redirects to the OAuth provider for authentication.",
		Tags:          []string{"OAuth"},
		DefaultStatus: http.StatusFound,
	}, deps.OAuthHandler.HandleInitiateLogin)

	huma.Register(api, huma.Operation{
		OperationID: "oauthCallback",
		Method:      "GET",
		Path:        authBase + "/oauth/callback/{provider}",
		Summary:     "OAuth callback",
		Description: "Handles the OAuth callback from the provider.",
		Tags:        []string{"OAuth"},
	}, deps.OAuthHandler.HandleCallback)

	huma.Register(api, huma.Operation{
		OperationID: "oauthCompleteWithEmail",
		Method:      "POST",
		Path:        authBase + "/oauth/complete",
		Summary:     "Complete OAuth with email",
		Description: "Completes OAuth flow when email is required.",
		Tags:        []string{"OAuth"},
	}, deps.OAuthHandler.HandleCompleteWithEmail)

	huma.Register(api, huma.Operation{
		OperationID: "initiateOAuthLink",
		Method:      "GET",
		Path:        authBase + "/oauth/link/{provider}",
		Summary:     "Initiate OAuth link",
		Description: "Initiates linking an OAuth provider to the authenticated account.",
		Tags:        []string{"OAuth"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.OAuthHandler.HandleInitiateLink)

	huma.Register(api, huma.Operation{
		OperationID: "oauthLinkCallback",
		Method:      "GET",
		Path:        authBase + "/oauth/link/callback/{provider}",
		Summary:     "OAuth link callback",
		Description: "Handles the OAuth callback when linking a provider.",
		Tags:        []string{"OAuth"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.OAuthHandler.HandleLinkCallback)

	huma.Register(api, huma.Operation{
		OperationID: "getOAuthIdentities",
		Method:      "GET",
		Path:        authBase + "/oauth/identities",
		Summary:     "List linked OAuth identities",
		Description: "Returns all OAuth providers linked to the authenticated account.",
		Tags:        []string{"OAuth"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.OAuthHandler.HandleGetIdentities)

	huma.Register(api, huma.Operation{
		OperationID: "unlinkOAuthProvider",
		Method:      "DELETE",
		Path:        authBase + "/oauth/identities/{provider}",
		Summary:     "Unlink OAuth provider",
		Description: "Unlinks an OAuth provider from the authenticated account.",
		Tags:        []string{"OAuth"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.OAuthHandler.HandleUnlink)
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
