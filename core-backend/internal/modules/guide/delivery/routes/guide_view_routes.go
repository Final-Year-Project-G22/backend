package routes

import (
	"github.com/danielgtaylor/huma/v2"
)

const (
	guideBase = "/api/v1/guides"
)

func RegisterGuideViewRoutes(api huma.API, deps RouteDependencies) {
	huma.Register(api, huma.Operation{
		OperationID: "listGuides",
		Method:      "GET",
		Path:        guideBase,
		Summary:     "List guides",
		Description: "Lists guides filtered by the user's business profile taxonomy (sector, tags, region, stage).",
		Tags:        []string{"Guides"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideViewHandler.HandleListGuides)

	huma.Register(api, huma.Operation{
		OperationID: "searchGuides",
		Method:      "GET",
		Path:        guideBase + "/search",
		Summary:     "Search guides",
		Description: "Searches guides by keyword with localized results.",
		Tags:        []string{"Guides"},
	}, deps.GuideViewHandler.HandleSearchGuides)

	huma.Register(api, huma.Operation{
		OperationID: "getRecentlyViewed",
		Method:      "GET",
		Path:        guideBase + "/recent",
		Summary:     "Get recently viewed guides",
		Description: "Retrieves the user's recently viewed guides with localized names.",
		Tags:        []string{"Guides"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideViewHandler.HandleGetRecentlyViewed)

	huma.Register(api, huma.Operation{
		OperationID: "getPersonalizedGuide",
		Method:      "GET",
		Path:        guideBase + "/{guideSlug}",
		Summary:     "Get personalized guide",
		Description: "Retrieves a guide with personalized step statuses based on user progress.",
		Tags:        []string{"Guides"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideViewHandler.HandleGetPersonalizedGuide)

	huma.Register(api, huma.Operation{
		OperationID: "getCurrentStep",
		Method:      "GET",
		Path:        guideBase + "/{guideSlug}/current-step",
		Summary:     "Get current step",
		Description: "Returns the next incomplete step in a guide for the user.",
		Tags:        []string{"Guides"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideViewHandler.HandleGetCurrentStep)

	huma.Register(api, huma.Operation{
		OperationID: "startStep",
		Method:      "POST",
		Path:        guideBase + "/steps/{stepId}/start",
		Summary:     "Start a step",
		Description: "Marks a step as in-progress for the user.",
		Tags:        []string{"Guides"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideViewHandler.HandleStartStep)

	huma.Register(api, huma.Operation{
		OperationID: "completeStep",
		Method:      "POST",
		Path:        guideBase + "/steps/{stepId}/complete",
		Summary:     "Complete a step",
		Description: "Marks a step as completed with optional documents and notes.",
		Tags:        []string{"Guides"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideViewHandler.HandleCompleteStep)

	huma.Register(api, huma.Operation{
		OperationID: "markStepIncomplete",
		Method:      "POST",
		Path:        guideBase + "/steps/{stepId}/mark-incomplete",
		Summary:     "Mark step as incomplete",
		Description: "Resets a step's status to in-progress.",
		Tags:        []string{"Guides"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideViewHandler.HandleMarkStepIncomplete)

	huma.Register(api, huma.Operation{
		OperationID: "skipOptionalStep",
		Method:      "POST",
		Path:        guideBase + "/steps/{stepId}/skip",
		Summary:     "Skip an optional step",
		Description: "Skips an optional step with an optional reason.",
		Tags:        []string{"Guides"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideViewHandler.HandleSkipOptionalStep)

	huma.Register(api, huma.Operation{
		OperationID: "updateProgress",
		Method:      "PATCH",
		Path:        guideBase + "/steps/{stepId}/progress",
		Summary:     "Update step progress",
		Description: "Updates progress details for a step including documents and notes.",
		Tags:        []string{"Guides"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideViewHandler.HandleUpdateProgress)

	huma.Register(api, huma.Operation{
		OperationID: "addBookmark",
		Method:      "POST",
		Path:        guideBase + "/steps/{stepId}/bookmark",
		Summary:     "Add bookmark",
		Description: "Bookmarks a step with an optional note.",
		Tags:        []string{"Guides"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideViewHandler.HandleAddBookmark)

	huma.Register(api, huma.Operation{
		OperationID: "updateBookmarkNote",
		Method:      "PATCH",
		Path:        guideBase + "/steps/{stepId}/bookmark",
		Summary:     "Update bookmark note",
		Description: "Updates the note on an existing bookmark.",
		Tags:        []string{"Guides"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideViewHandler.HandleUpdateBookmarkNote)

	huma.Register(api, huma.Operation{
		OperationID: "removeBookmark",
		Method:      "DELETE",
		Path:        guideBase + "/steps/{stepId}/bookmark",
		Summary:     "Remove bookmark",
		Description: "Removes a bookmark from a step.",
		Tags:        []string{"Guides"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideViewHandler.HandleRemoveBookmark)

	huma.Register(api, huma.Operation{
		OperationID: "listBookmarks",
		Method:      "GET",
		Path:        guideBase + "/bookmarks",
		Summary:     "List bookmarks",
		Description: "Lists all bookmarks for the user.",
		Tags:        []string{"Guides"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideViewHandler.HandleListBookmarks)
}
