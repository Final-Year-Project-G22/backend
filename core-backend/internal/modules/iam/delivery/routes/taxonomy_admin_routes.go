package routes

import (
	sharedmiddleware "github.com/Final-Year-Project-G22/backend/core/internal/shared/middleware"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/permissions"
	"github.com/danielgtaylor/huma/v2"
)

func RegisterTaxonomyAdminRoutes(api huma.API, deps AdminRouteDependencies) {
	readPermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, deps.RoleAssignmentUsecase, permissions.IAMRead, []string{"super_admin"})
	writePermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, deps.RoleAssignmentUsecase, permissions.IAMWrite, []string{"super_admin"})
	updatePermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, deps.RoleAssignmentUsecase, permissions.IAMUpdate, []string{"super_admin"})
	deletePermissionMiddleware := sharedmiddleware.PermissionMiddleware(api, deps.RoleAssignmentUsecase, permissions.IAMDelete, []string{"super_admin"})

	huma.Register(api, huma.Operation{
		OperationID: "listSectors",
		Method:      "GET",
		Path:        adminBase + "/sectors",
		Summary:     "List sectors",
		Tags:        []string{"Admin - Taxonomy"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, readPermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.TaxonomyAdminHandler.HandleListSectors)

	huma.Register(api, huma.Operation{
		OperationID: "getSector",
		Method:      "GET",
		Path:        adminBase + "/sectors/{id}",
		Summary:     "Get sector",
		Tags:        []string{"Admin - Taxonomy"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, readPermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.TaxonomyAdminHandler.HandleGetSector)

	huma.Register(api, huma.Operation{
		OperationID: "createSector",
		Method:      "POST",
		Path:        adminBase + "/sectors",
		Summary:     "Create sector",
		Tags:        []string{"Admin - Taxonomy"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, writePermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.TaxonomyAdminHandler.HandleCreateSector)

	huma.Register(api, huma.Operation{
		OperationID: "updateSector",
		Method:      "PUT",
		Path:        adminBase + "/sectors/{id}",
		Summary:     "Update sector",
		Tags:        []string{"Admin - Taxonomy"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, updatePermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.TaxonomyAdminHandler.HandleUpdateSector)

	huma.Register(api, huma.Operation{
		OperationID: "deleteSector",
		Method:      "DELETE",
		Path:        adminBase + "/sectors/{id}",
		Summary:     "Delete sector",
		Tags:        []string{"Admin - Taxonomy"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deletePermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.TaxonomyAdminHandler.HandleDeleteSector)

	huma.Register(api, huma.Operation{
		OperationID: "listTags",
		Method:      "GET",
		Path:        adminBase + "/tags",
		Summary:     "List tags",
		Tags:        []string{"Admin - Taxonomy"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, readPermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.TaxonomyAdminHandler.HandleListTags)

	huma.Register(api, huma.Operation{
		OperationID: "getTag",
		Method:      "GET",
		Path:        adminBase + "/tags/{id}",
		Summary:     "Get tag",
		Tags:        []string{"Admin - Taxonomy"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, readPermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.TaxonomyAdminHandler.HandleGetTag)

	huma.Register(api, huma.Operation{
		OperationID: "createTag",
		Method:      "POST",
		Path:        adminBase + "/tags",
		Summary:     "Create tag",
		Tags:        []string{"Admin - Taxonomy"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, writePermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.TaxonomyAdminHandler.HandleCreateTag)

	huma.Register(api, huma.Operation{
		OperationID: "updateTag",
		Method:      "PUT",
		Path:        adminBase + "/tags/{id}",
		Summary:     "Update tag",
		Tags:        []string{"Admin - Taxonomy"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, updatePermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.TaxonomyAdminHandler.HandleUpdateTag)

	huma.Register(api, huma.Operation{
		OperationID: "deleteTag",
		Method:      "DELETE",
		Path:        adminBase + "/tags/{id}",
		Summary:     "Delete tag",
		Tags:        []string{"Admin - Taxonomy"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deletePermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.TaxonomyAdminHandler.HandleDeleteTag)
}
