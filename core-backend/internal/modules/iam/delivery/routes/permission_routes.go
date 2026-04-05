package routes

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/dto"
	"github.com/danielgtaylor/huma/v2"
)

const permissionsBase = apiV1Base + "/permissions"

type PermissionRouteDependencies struct {
	PermissionHandler        PermissionHandler
	AuthMiddleware           func(huma.Context, func(huma.Context))
	AccountStatusMiddleware  func(huma.Context, func(huma.Context))
	ReadPermissionMiddleware func(huma.Context, func(huma.Context))
}

type PermissionHandler interface {
	HandleListPermissions(ctx context.Context, input *dto.ListPermissionsInput) (*dto.ListPermissionsOutput, error)
}

func RegisterPermissionRoutes(api huma.API, deps PermissionRouteDependencies) {
	huma.Register(api, huma.Operation{
		OperationID: "listPermissions",
		Method:      "GET",
		Path:        permissionsBase,
		Summary:     "List permissions",
		Description: "Returns permissions with optional filtering by module and code.",
		Tags:        []string{"Permissions"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.PermissionHandler.HandleListPermissions)
}
