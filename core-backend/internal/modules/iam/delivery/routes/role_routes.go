package routes

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	sharedmiddleware "github.com/Final-Year-Project-G22/backend/core/internal/shared/middleware"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/permissions"
	"github.com/danielgtaylor/huma/v2"
)

const rolesBase = apiV1Base + "/roles"

type RoleHandler interface {
	HandleGetRole(ctx context.Context, input *dto.GetRoleInput) (*dto.GetRoleOutput, error)
	HandleListRoles(ctx context.Context, input *dto.ListRolesInput) (*dto.ListRolesOutput, error)
	HandleCreateRole(ctx context.Context, input *dto.CreateRoleInput) (*dto.CreateRoleOutput, error)
	HandleUpdateRole(ctx context.Context, input *dto.UpdateRoleInput) (*dto.UpdateRoleOutput, error)
	HandleDeleteRole(ctx context.Context, input *dto.DeleteRoleInput) (*dto.DeleteRoleOutput, error)
}

type RoleRouteDependencies struct {
	RoleHandler             RoleHandler
	AuthMiddleware          func(huma.Context, func(huma.Context))
	AccountStatusMiddleware func(huma.Context, func(huma.Context))
	RoleAssignmentUsecase   usecase.RoleAssignmentUsecase
}

func RegisterRoleRoutes(api huma.API, deps RoleRouteDependencies) {
	readPermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, deps.RoleAssignmentUsecase, permissions.IAMRead, nil)
	writePermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, deps.RoleAssignmentUsecase, permissions.IAMWrite, nil)
	updatePermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, deps.RoleAssignmentUsecase, permissions.IAMUpdate, nil)
	deletePermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, deps.RoleAssignmentUsecase, permissions.IAMDelete, nil)

	huma.Register(api, huma.Operation{
		OperationID: "getRole",
		Method:      "GET",
		Path:        rolesBase + "/{roleId}",
		Summary:     "Get role",
		Description: "Returns role details and its permissions.",
		Tags:        []string{"Roles"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, readPermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.RoleHandler.HandleGetRole)

	huma.Register(api, huma.Operation{
		OperationID: "listRoles",
		Method:      "GET",
		Path:        rolesBase,
		Summary:     "List roles",
		Description: "Returns all roles.",
		Tags:        []string{"Roles"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, readPermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.RoleHandler.HandleListRoles)

	huma.Register(api, huma.Operation{
		OperationID: "createRole",
		Method:      "POST",
		Path:        rolesBase,
		Summary:     "Create role",
		Description: "Creates a custom role and assigns permissions.",
		Tags:        []string{"Roles"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, writePermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.RoleHandler.HandleCreateRole)

	huma.Register(api, huma.Operation{
		OperationID: "updateRole",
		Method:      "PUT",
		Path:        rolesBase + "/{roleId}",
		Summary:     "Update role",
		Description: "Updates role details and replaces permissions.",
		Tags:        []string{"Roles"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, updatePermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.RoleHandler.HandleUpdateRole)

	huma.Register(api, huma.Operation{
		OperationID: "deleteRole",
		Method:      "DELETE",
		Path:        rolesBase + "/{roleId}",
		Summary:     "Delete role",
		Description: "Permanently deletes a role if it is mutable.",
		Tags:        []string{"Roles"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deletePermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.RoleHandler.HandleDeleteRole)
}
