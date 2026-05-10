package routes

import "github.com/danielgtaylor/huma/v2"

const communityBase = "/api/v1/community"

func RegisterCommunityRoutes(api huma.API, deps RouteDependencies) {
	huma.Register(api, huma.Operation{
		OperationID: "listCommunityCategories",
		Method:      "GET",
		Path:        communityBase + "/categories",
		Summary:     "List community categories",
		Description: "Lists active community categories.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleListCategories)

	huma.Register(api, huma.Operation{
		OperationID: "getCommunityCategory",
		Method:      "GET",
		Path:        communityBase + "/categories/{id}",
		Summary:     "Get community category",
		Description: "Retrieves community category details.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleGetCategory)

	huma.Register(api, huma.Operation{
		OperationID: "listCommunityThreads",
		Method:      "GET",
		Path:        communityBase + "/threads",
		Summary:     "List discussion threads",
		Description: "Lists discussion threads filtered by user's business profile taxonomy.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleListThreads)

	huma.Register(api, huma.Operation{
		OperationID: "searchCommunityThreads",
		Method:      "GET",
		Path:        communityBase + "/threads/search",
		Summary:     "Search discussion threads",
		Description: "Searches threads by keyword, filtered by user's business profile taxonomy.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleSearchThreads)

	huma.Register(api, huma.Operation{
		OperationID: "getCommunityThread",
		Method:      "GET",
		Path:        communityBase + "/threads/{id}",
		Summary:     "Get discussion thread",
		Description: "Retrieves thread details.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleGetThread)

	huma.Register(api, huma.Operation{
		OperationID: "listCommunityPosts",
		Method:      "GET",
		Path:        communityBase + "/threads/{id}/posts",
		Summary:     "List thread posts",
		Description: "Lists posts in a thread.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleListPosts)

	huma.Register(api, huma.Operation{
		OperationID: "createCommunityThread",
		Method:      "POST",
		Path:        communityBase + "/threads",
		Summary:     "Create discussion thread",
		Description: "Creates a thread with an initial post.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleCreateThread)

	huma.Register(api, huma.Operation{
		OperationID: "createCommunityPost",
		Method:      "POST",
		Path:        communityBase + "/threads/{id}/posts",
		Summary:     "Create thread post",
		Description: "Creates a top-level post in a thread.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleCreatePost)

	huma.Register(api, huma.Operation{
		OperationID: "replyCommunityPost",
		Method:      "POST",
		Path:        communityBase + "/threads/{id}/posts/{postId}/reply",
		Summary:     "Reply to post",
		Description: "Replies to a post in a thread.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleReplyPost)

	huma.Register(api, huma.Operation{
		OperationID: "updateCommunityPost",
		Method:      "PATCH",
		Path:        communityBase + "/posts/{id}",
		Summary:     "Update post",
		Description: "Updates a post.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleUpdatePost)

	huma.Register(api, huma.Operation{
		OperationID: "deleteCommunityPost",
		Method:      "DELETE",
		Path:        communityBase + "/posts/{id}",
		Summary:     "Delete post",
		Description: "Deletes a post.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleDeletePost)

	huma.Register(api, huma.Operation{
		OperationID: "markCommunitySolution",
		Method:      "POST",
		Path:        communityBase + "/threads/{id}/solution/{postId}",
		Summary:     "Mark accepted solution",
		Description: "Marks a post as the accepted solution.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleMarkSolution)

	huma.Register(api, huma.Operation{
		OperationID: "followCommunityThread",
		Method:      "POST",
		Path:        communityBase + "/threads/{id}/follow",
		Summary:     "Follow thread",
		Description: "Follows a discussion thread.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleFollowThread)

	huma.Register(api, huma.Operation{
		OperationID: "unfollowCommunityThread",
		Method:      "DELETE",
		Path:        communityBase + "/threads/{id}/follow",
		Summary:     "Unfollow thread",
		Description: "Unfollows a discussion thread.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleUnfollowThread)

	huma.Register(api, huma.Operation{
		OperationID: "followCommunityCategory",
		Method:      "POST",
		Path:        communityBase + "/categories/{id}/follow",
		Summary:     "Follow category",
		Description: "Follows a community category.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleFollowCategory)

	huma.Register(api, huma.Operation{
		OperationID: "unfollowCommunityCategory",
		Method:      "DELETE",
		Path:        communityBase + "/categories/{id}/follow",
		Summary:     "Unfollow category",
		Description: "Unfollows a community category.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleUnfollowCategory)

	huma.Register(api, huma.Operation{
		OperationID: "reportThread",
		Method:      "POST",
		Path:        communityBase + "/threads/{id}/reports",
		Summary:     "Report thread",
		Description: "Reports a discussion thread.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleReportThread)

	huma.Register(api, huma.Operation{
		OperationID: "reportPost",
		Method:      "POST",
		Path:        communityBase + "/threads/{id}/posts/{postId}/reports",
		Summary:     "Report post",
		Description: "Reports a discussion post.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleReportPost)

	huma.Register(api, huma.Operation{
		OperationID: "reportUser",
		Method:      "POST",
		Path:        communityBase + "/threads/{id}/reports/user",
		Summary:     "Report user",
		Description: "Reports a user in a thread.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleReportUser)

	huma.Register(api, huma.Operation{
		OperationID: "listFollowedThreads",
		Method:      "GET",
		Path:        communityBase + "/follows/threads",
		Summary:     "List followed threads",
		Description: "Lists threads followed by the current user.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleListFollowedThreads)

	huma.Register(api, huma.Operation{
		OperationID: "listFollowedCategories",
		Method:      "GET",
		Path:        communityBase + "/follows/categories",
		Summary:     "List followed categories",
		Description: "Lists categories followed by the current user.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleListFollowedCategories)

	huma.Register(api, huma.Operation{
		OperationID: "blockCommunityUser",
		Method:      "POST",
		Path:        communityBase + "/threads/{id}/blocks",
		Summary:     "Block user in thread",
		Description: "Blocks a user from a thread (thread author only).",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleBlockUser)

	huma.Register(api, huma.Operation{
		OperationID: "uploadAttachments",
		Method:      "POST",
		Path:        communityBase + "/attachments",
		Summary:     "Upload attachments",
		Description: "Uploads one or more attachments for later use when creating/updating posts.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleUploadAttachments)

	huma.Register(api, huma.Operation{
		OperationID: "deleteOrphanAttachment",
		Method:      "DELETE",
		Path:        communityBase + "/attachments/{id}",
		Summary:     "Delete orphan attachment",
		Description: "Deletes a pending attachment that was uploaded but not yet linked to a post.",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleDeleteOrphanAttachment)

	huma.Register(api, huma.Operation{
		OperationID: "unblockCommunityUser",
		Method:      "DELETE",
		Path:        communityBase + "/threads/{id}/blocks/{accountId}",
		Summary:     "Unblock user in thread",
		Description: "Unblocks a user from a thread (thread author only).",
		Tags:        []string{"Community"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.CommunityHandler.HandleUnblockUser)
}
