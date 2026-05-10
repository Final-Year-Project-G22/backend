package routes

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	sharedmiddleware "github.com/Final-Year-Project-G22/backend/core/internal/shared/middleware"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/permissions"
	"github.com/danielgtaylor/huma/v2"
)

const permissionsBase = apiV1Base + "/permissions"

type PermissionRouteDependencies struct {
	PermissionHandler       PermissionHandler
	AuthMiddleware          func(huma.Context, func(huma.Context))
	AccountStatusMiddleware func(huma.Context, func(huma.Context))
	RoleAssignmentUsecase   usecase.RoleAssignmentUsecase
}

type PermissionHandler interface {
	HandleListPermissions(ctx context.Context, input *dto.ListPermissionsInput) (*dto.ListPermissionsOutput, error)
}

func RegisterPermissionRoutes(api huma.API, deps PermissionRouteDependencies) {
	readPermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, deps.RoleAssignmentUsecase, permissions.IAMRead, nil)

	huma.Register(api, huma.Operation{
		OperationID: "listPermissions",
		Method:      "GET",
		Path:        permissionsBase,
		Summary:     "List permissions",
		Description: "Returns permissions with optional filtering by module and code.",
		Tags:        []string{"Permissions"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, readPermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.PermissionHandler.HandleListPermissions)
}
