package routes

import (
	"github.com/danielgtaylor/huma/v2"
)

const (
	adminGuideBase = "/api/v1/admin/guides"
)

func RegisterGuideAdminRoutes(api huma.API, deps RouteDependencies) {
	huma.Register(api, huma.Operation{
		OperationID: "listGuidesAdmin",
		Method:      "GET",
		Path:        adminGuideBase,
		Summary:     "List guides",
		Description: "Lists guides for admin management with pagination and filters.",
		Tags:        []string{"Admin - Guides"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleListGuides)

	huma.Register(api, huma.Operation{
		OperationID: "getGuideAdmin",
		Method:      "GET",
		Path:        adminGuideBase + "/{id}",
		Summary:     "Get guide detail",
		Description: "Retrieves guide detail for admin editor.",
		Tags:        []string{"Admin - Guides"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleGetGuideAdmin)

	huma.Register(api, huma.Operation{
		OperationID: "listGuideStepsAdmin",
		Method:      "GET",
		Path:        adminGuideBase + "/{id}/steps",
		Summary:     "List guide steps",
		Description: "Lists steps of a guide for admin editor.",
		Tags:        []string{"Admin - Steps"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleListGuideStepsAdmin)

	huma.Register(api, huma.Operation{
		OperationID: "createGuide",
		Method:      "POST",
		Path:        adminGuideBase,
		Summary:     "Create guide",
		Description: "Creates a new guide with optional translations and conditions.",
		Tags:        []string{"Admin - Guides"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleCreateGuide)

	huma.Register(api, huma.Operation{
		OperationID: "updateGuide",
		Method:      "PUT",
		Path:        adminGuideBase + "/{id}",
		Summary:     "Update guide",
		Description: "Updates an existing guide.",
		Tags:        []string{"Admin - Guides"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleUpdateGuide)

	huma.Register(api, huma.Operation{
		OperationID: "deleteGuide",
		Method:      "DELETE",
		Path:        adminGuideBase + "/{id}",
		Summary:     "Delete guide",
		Description: "Deletes a guide.",
		Tags:        []string{"Admin - Guides"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleDeleteGuide)

	huma.Register(api, huma.Operation{
		OperationID: "addGuideCondition",
		Method:      "POST",
		Path:        adminGuideBase + "/{id}/conditions",
		Summary:     "Add guide condition",
		Description: "Adds a visibility condition to a guide.",
		Tags:        []string{"Admin - Guides"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleAddGuideCondition)

	huma.Register(api, huma.Operation{
		OperationID: "removeGuideCondition",
		Method:      "DELETE",
		Path:        adminGuideBase + "/conditions/{condId}",
		Summary:     "Remove guide condition",
		Description: "Removes a visibility condition from a guide.",
		Tags:        []string{"Admin - Guides"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleRemoveGuideCondition)

	huma.Register(api, huma.Operation{
		OperationID: "setGuideTranslations",
		Method:      "PUT",
		Path:        adminGuideBase + "/{id}/translations",
		Summary:     "Set guide translations",
		Description: "Replaces all translations for a guide.",
		Tags:        []string{"Admin - Guides"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleSetGuideTranslations)

	huma.Register(api, huma.Operation{
		OperationID: "createStep",
		Method:      "POST",
		Path:        adminGuideBase + "/steps",
		Summary:     "Create step",
		Description: "Creates a new guide step with translations, conditions, and dependencies.",
		Tags:        []string{"Admin - Steps"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleCreateStep)

	huma.Register(api, huma.Operation{
		OperationID: "updateStep",
		Method:      "PUT",
		Path:        adminGuideBase + "/steps/{id}",
		Summary:     "Update step",
		Description: "Updates an existing guide step.",
		Tags:        []string{"Admin - Steps"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleUpdateStep)

	huma.Register(api, huma.Operation{
		OperationID: "deleteStep",
		Method:      "DELETE",
		Path:        adminGuideBase + "/steps/{id}",
		Summary:     "Delete step",
		Description: "Deletes a guide step.",
		Tags:        []string{"Admin - Steps"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleDeleteStep)

	huma.Register(api, huma.Operation{
		OperationID: "reorderSteps",
		Method:      "PUT",
		Path:        adminGuideBase + "/steps/reorder",
		Summary:     "Reorder steps",
		Description: "Reorders the steps of a guide.",
		Tags:        []string{"Admin - Steps"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleReorderSteps)

	huma.Register(api, huma.Operation{
		OperationID: "addStepCondition",
		Method:      "POST",
		Path:        adminGuideBase + "/steps/{id}/conditions",
		Summary:     "Add step condition",
		Description: "Adds a visibility condition to a step.",
		Tags:        []string{"Admin - Steps"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleAddStepCondition)

	huma.Register(api, huma.Operation{
		OperationID: "removeStepCondition",
		Method:      "DELETE",
		Path:        adminGuideBase + "/steps/conditions/{condId}",
		Summary:     "Remove step condition",
		Description: "Removes a visibility condition from a step.",
		Tags:        []string{"Admin - Steps"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleRemoveStepCondition)

	huma.Register(api, huma.Operation{
		OperationID: "addStepDependency",
		Method:      "POST",
		Path:        adminGuideBase + "/steps/{id}/dependencies",
		Summary:     "Add step dependency",
		Description: "Adds a dependency to a step.",
		Tags:        []string{"Admin - Steps"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleAddStepDependency)

	huma.Register(api, huma.Operation{
		OperationID: "removeStepDependency",
		Method:      "DELETE",
		Path:        adminGuideBase + "/steps/dependencies/{depId}",
		Summary:     "Remove step dependency",
		Description: "Removes a dependency from a step.",
		Tags:        []string{"Admin - Steps"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleRemoveStepDependency)

	huma.Register(api, huma.Operation{
		OperationID: "setStepTranslations",
		Method:      "PUT",
		Path:        adminGuideBase + "/steps/{id}/translations",
		Summary:     "Set step translations",
		Description: "Replaces all translations for a step.",
		Tags:        []string{"Admin - Steps"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleSetStepTranslations)

	huma.Register(api, huma.Operation{
		OperationID: "getStepVersions",
		Method:      "GET",
		Path:        adminGuideBase + "/steps/{id}/versions",
		Summary:     "Get step versions",
		Description: "Lists all versions of a step.",
		Tags:        []string{"Admin - Steps"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleGetStepVersions)

	huma.Register(api, huma.Operation{
		OperationID: "revertStepToVersion",
		Method:      "POST",
		Path:        adminGuideBase + "/steps/{id}/versions/{version}/revert",
		Summary:     "Revert step to version",
		Description: "Reverts a step to a previous version (not yet supported).",
		Tags:        []string{"Admin - Steps"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleRevertStepToVersion)

	huma.Register(api, huma.Operation{
		OperationID: "invalidateUserJourney",
		Method:      "DELETE",
		Path:        adminGuideBase + "/journeys/users/{userId}",
		Summary:     "Invalidate user journey",
		Description: "Invalidates a specific user's journey for a guide.",
		Tags:        []string{"Admin - Journeys"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleInvalidateUserJourney)

	huma.Register(api, huma.Operation{
		OperationID: "invalidateAllJourneys",
		Method:      "DELETE",
		Path:        adminGuideBase + "/journeys",
		Summary:     "Invalidate all journeys",
		Description: "Invalidates all journeys for a guide.",
		Tags:        []string{"Admin - Journeys"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.GuideAdminHandler.HandleInvalidateAllJourneys)
}
