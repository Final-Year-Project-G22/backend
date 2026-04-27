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

	huma.Register(api, huma.Operation{
		OperationID: "adminListThreadBlockedUsers",
		Method:      "GET",
		Path:        communityAdminBase + "/threads/{id}/blocks",
		Summary:     "List blocked users in thread (admin)",
		Description: "Lists blocked users for a specific thread.",
		Tags:        []string{"Admin - Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityAdminHandler.HandleListBlockedUsers)

	huma.Register(api, huma.Operation{
		OperationID: "adminListAllBlockedUsers",
		Method:      "GET",
		Path:        communityAdminBase + "/blocks",
		Summary:     "List all blocked users (admin)",
		Description: "Lists all blocked users across all threads.",
		Tags:        []string{"Admin - Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityAdminHandler.HandleListAllBlockedUsers)

	huma.Register(api, huma.Operation{
		OperationID: "adminListThreadReports",
		Method:      "GET",
		Path:        communityAdminBase + "/reports/threads",
		Summary:     "List thread reports (admin)",
		Description: "Lists thread reports with optional status filter.",
		Tags:        []string{"Admin - Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityAdminHandler.HandleListThreadReports)

	huma.Register(api, huma.Operation{
		OperationID: "adminGetThreadReport",
		Method:      "GET",
		Path:        communityAdminBase + "/reports/threads/{id}",
		Summary:     "Get thread report (admin)",
		Description: "Gets a thread report with reported content details.",
		Tags:        []string{"Admin - Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityAdminHandler.HandleGetThreadReport)

	huma.Register(api, huma.Operation{
		OperationID: "adminUpdateThreadReportStatus",
		Method:      "PATCH",
		Path:        communityAdminBase + "/reports/threads/{id}/status",
		Summary:     "Update thread report status (admin)",
		Description: "Updates a thread report status.",
		Tags:        []string{"Admin - Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.UpdatePermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityAdminHandler.HandleUpdateThreadReportStatus)

	huma.Register(api, huma.Operation{
		OperationID: "adminDeleteReportedThread",
		Method:      "DELETE",
		Path:        communityAdminBase + "/reports/threads/{id}",
		Summary:     "Delete reported thread (admin)",
		Description: "Deletes a reported thread and resolves the report.",
		Tags:        []string{"Admin - Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.DeletePermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityAdminHandler.HandleDeleteReportedThread)

	huma.Register(api, huma.Operation{
		OperationID: "adminListPostReports",
		Method:      "GET",
		Path:        communityAdminBase + "/reports/posts",
		Summary:     "List post reports (admin)",
		Description: "Lists post reports with optional status filter.",
		Tags:        []string{"Admin - Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityAdminHandler.HandleListPostReports)

	huma.Register(api, huma.Operation{
		OperationID: "adminGetPostReport",
		Method:      "GET",
		Path:        communityAdminBase + "/reports/posts/{id}",
		Summary:     "Get post report (admin)",
		Description: "Gets a post report with reported content details.",
		Tags:        []string{"Admin - Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityAdminHandler.HandleGetPostReport)

	huma.Register(api, huma.Operation{
		OperationID: "adminUpdatePostReportStatus",
		Method:      "PATCH",
		Path:        communityAdminBase + "/reports/posts/{id}/status",
		Summary:     "Update post report status (admin)",
		Description: "Updates a post report status.",
		Tags:        []string{"Admin - Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.UpdatePermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityAdminHandler.HandleUpdatePostReportStatus)

	huma.Register(api, huma.Operation{
		OperationID: "adminDeleteReportedPost",
		Method:      "DELETE",
		Path:        communityAdminBase + "/reports/posts/{id}",
		Summary:     "Delete reported post (admin)",
		Description: "Deletes a reported post and resolves the report.",
		Tags:        []string{"Admin - Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.DeletePermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityAdminHandler.HandleDeleteReportedPost)

	huma.Register(api, huma.Operation{
		OperationID: "adminListUserReports",
		Method:      "GET",
		Path:        communityAdminBase + "/reports/users",
		Summary:     "List user reports (admin)",
		Description: "Lists user reports with optional status filter.",
		Tags:        []string{"Admin - Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityAdminHandler.HandleListUserReports)

	huma.Register(api, huma.Operation{
		OperationID: "adminGetUserReport",
		Method:      "GET",
		Path:        communityAdminBase + "/reports/users/{id}",
		Summary:     "Get user report (admin)",
		Description: "Gets a user report with reported user details.",
		Tags:        []string{"Admin - Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.ReadPermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityAdminHandler.HandleGetUserReport)

	huma.Register(api, huma.Operation{
		OperationID: "adminUpdateUserReportStatus",
		Method:      "PATCH",
		Path:        communityAdminBase + "/reports/users/{id}/status",
		Summary:     "Update user report status (admin)",
		Description: "Updates a user report status.",
		Tags:        []string{"Admin - Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.UpdatePermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityAdminHandler.HandleUpdateUserReportStatus)

	huma.Register(api, huma.Operation{
		OperationID: "adminBlockReportedUser",
		Method:      "POST",
		Path:        communityAdminBase + "/reports/users/{id}/block",
		Summary:     "Block reported user (admin)",
		Description: "Blocks the reported user from the thread and resolves the report.",
		Tags:        []string{"Admin - Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, deps.UpdatePermissionMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityAdminHandler.HandleBlockReportedUser)
}
