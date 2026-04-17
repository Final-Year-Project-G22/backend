package routes

import "github.com/danielgtaylor/huma/v2"

const communityAdminBase = "/api/v1/admin/community"

func RegisterCommunityAdminRoutes(api huma.API, deps RouteDependencies) {
	huma.Register(api, huma.Operation{
		OperationID: "adminListCommunityCategories",
		Method:      "GET",
		Path:        communityAdminBase + "/categories",
		Summary:     "List community categories (admin)",
		Description: "Lists community categories including inactive ones.",
		Tags:        []string{"Admin - Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityAdminHandler.HandleListCategories)

	huma.Register(api, huma.Operation{
		OperationID: "adminCreateCommunityCategory",
		Method:      "POST",
		Path:        communityAdminBase + "/categories",
		Summary:     "Create community category",
		Description: "Creates a community category.",
		Tags:        []string{"Admin - Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.WritePermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityAdminHandler.HandleCreateCategory)

	huma.Register(api, huma.Operation{
		OperationID: "adminUpdateCommunityCategory",
		Method:      "PUT",
		Path:        communityAdminBase + "/categories/{id}",
		Summary:     "Update community category",
		Description: "Updates a community category.",
		Tags:        []string{"Admin - Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.UpdatePermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityAdminHandler.HandleUpdateCategory)

	huma.Register(api, huma.Operation{
		OperationID: "adminDeleteCommunityCategory",
		Method:      "DELETE",
		Path:        communityAdminBase + "/categories/{id}",
		Summary:     "Delete community category",
		Description: "Deletes a community category.",
		Tags:        []string{"Admin - Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.DeletePermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityAdminHandler.HandleDeleteCategory)

	huma.Register(api, huma.Operation{
		OperationID: "adminBlockCommunityUser",
		Method:      "POST",
		Path:        communityAdminBase + "/threads/{id}/blocks",
		Summary:     "Block user in thread (admin)",
		Description: "Blocks a user from a thread as admin.",
		Tags:        []string{"Admin - Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.UpdatePermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityAdminHandler.HandleBlockUser)

	huma.Register(api, huma.Operation{
		OperationID: "adminUnblockCommunityUser",
		Method:      "DELETE",
		Path:        communityAdminBase + "/threads/{id}/blocks/{accountId}",
		Summary:     "Unblock user in thread (admin)",
		Description: "Unblocks a user from a thread as admin.",
		Tags:        []string{"Admin - Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.UpdatePermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityAdminHandler.HandleUnblockUser)
}
