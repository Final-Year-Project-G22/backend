package routes

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/handler"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	sharedmiddleware "github.com/Final-Year-Project-G22/backend/core/internal/shared/middleware"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/permissions"
	"github.com/danielgtaylor/huma/v2"
)

const adminBase = apiV1Base + "/admin"

type AdminRouteDependencies struct {
	AdminHandler            *handler.AdminHandler
	TaxonomyAdminHandler    *handler.TaxonomyAdminHandler
	AuthMiddleware          func(huma.Context, func(huma.Context))
	AccountStatusMiddleware func(huma.Context, func(huma.Context))
	RoleAssignmentUsecase   usecase.RoleAssignmentUsecase
}

func RegisterAdminRoutes(api huma.API, deps AdminRouteDependencies) {
	adminListMiddleware := sharedmiddleware.PermissionMiddleware(api, deps.RoleAssignmentUsecase, permissions.AdminList, []string{"super_admin"})
	adminStatusUpdateMiddleware := sharedmiddleware.PermissionMiddleware(api, deps.RoleAssignmentUsecase, permissions.AdminStatusUpdate, []string{"super_admin"})
	adminResetPasswordMiddleware := sharedmiddleware.PermissionMiddleware(api, deps.RoleAssignmentUsecase, permissions.AdminResetPassword, []string{"super_admin"})

	huma.Register(api, huma.Operation{
		OperationID: "listAdmins",
		Method:      "GET",
		Path:        adminBase + "/accounts",
		Summary:     "List admin accounts",
		Description: "Returns a paginated list of admin accounts filtered by search, status, and role.",
		Tags:        []string{"Admin Management"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, adminListMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.AdminHandler.HandleListAdmins)

	huma.Register(api, huma.Operation{
		OperationID: "updateAdminStatus",
		Method:      "PATCH",
		Path:        adminBase + "/accounts/{accountId}/status",
		Summary:     "Update admin account status",
		Description: "Updates the status of an admin account (active, locked, suspended).",
		Tags:        []string{"Admin Management"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, adminStatusUpdateMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.AdminHandler.HandleUpdateAdminStatus)

	huma.Register(api, huma.Operation{
		OperationID: "resetAdminPassword",
		Method:      "POST",
		Path:        adminBase + "/accounts/{accountId}/reset-password",
		Summary:     "Trigger admin password reset",
		Description: "Triggers a password reset for an admin account by sending an OTP to their email.",
		Tags:        []string{"Admin Management"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware, deps.AccountStatusMiddleware, adminResetPasswordMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.AdminHandler.HandleResetAdminPassword)

	huma.Register(api, huma.Operation{
		OperationID: "completeAdminPasswordReset",
		Method:      "POST",
		Path:        authBase + "/admin/reset-password",
		Summary:     "Complete admin password reset",
		Description: "Validates the reset token and sets a new password for the admin account.",
		Tags:        []string{"Authentication"},
	}, deps.AdminHandler.HandleCompleteAdminPasswordReset)

	RegisterTaxonomyAdminRoutes(api, deps)
}
