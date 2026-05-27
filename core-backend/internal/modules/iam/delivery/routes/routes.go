package routes

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/handler"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	"github.com/danielgtaylor/huma/v2"
)

type RouteDependencies struct {
	AuthHandler                 *handler.AuthHandler
	AdminHandler                *handler.AdminHandler
	DashboardHandler            *handler.DashboardHandler
	TaxonomyAdminHandler        *handler.TaxonomyAdminHandler
	TaxonomyHandler             *handler.TaxonomyHandler
	BusinessProfileHandler      *handler.BusinessProfileHandler
	BusinessProfileImageHandler *handler.BusinessProfileImageHandler
	PermissionHandler           *handler.PermissionHandler
	RoleHandler                 *handler.RoleHandler
	UserHandler                 *handler.UserHandler
	ImageHandler                *handler.ImageHandler
	OAuthHandler                *handler.OAuthHandler
	NotificationPrefHandler     *handler.NotificationPreferenceHandler
	AccountPrefHandler          *handler.AccountPreferenceHandler
	AuthMiddleware              func(huma.Context, func(huma.Context))
	AccountStatusMiddleware     func(huma.Context, func(huma.Context))
	RoleAssignmentUsecase       usecase.RoleAssignmentUsecase
}

func RegisterRoutes(api huma.API, deps RouteDependencies) {
	RegisterAuthRoutes(api, AuthRouteDependencies{
		AuthHandler:             deps.AuthHandler,
		AdminHandler:            deps.AdminHandler,
		OAuthHandler:            deps.OAuthHandler,
		AuthMiddleware:          deps.AuthMiddleware,
		AccountStatusMiddleware: deps.AccountStatusMiddleware,
		RoleAssignmentUsecase:   deps.RoleAssignmentUsecase,
	})
	RegisterUserRoutes(api, UserRouteDependencies{
		ImageHandler:            deps.ImageHandler,
		AuthMiddleware:          deps.AuthMiddleware,
		UserHandler:             deps.UserHandler,
		AccountStatusMiddleware: deps.AccountStatusMiddleware,
	})
	RegisterPermissionRoutes(api, PermissionRouteDependencies{
		PermissionHandler:       deps.PermissionHandler,
		AuthMiddleware:          deps.AuthMiddleware,
		AccountStatusMiddleware: deps.AccountStatusMiddleware,
		RoleAssignmentUsecase:   deps.RoleAssignmentUsecase,
	})
	RegisterRoleRoutes(api, RoleRouteDependencies{
		RoleHandler:             deps.RoleHandler,
		AuthMiddleware:          deps.AuthMiddleware,
		AccountStatusMiddleware: deps.AccountStatusMiddleware,
		RoleAssignmentUsecase:   deps.RoleAssignmentUsecase,
	})
	RegisterAdminRoutes(api, AdminRouteDependencies{
		AdminHandler:            deps.AdminHandler,
		TaxonomyAdminHandler:    deps.TaxonomyAdminHandler,
		AuthMiddleware:          deps.AuthMiddleware,
		AccountStatusMiddleware: deps.AccountStatusMiddleware,
		RoleAssignmentUsecase:   deps.RoleAssignmentUsecase,
	}, deps.DashboardHandler)
	RegisterTaxonomyRoutes(api, TaxonomyRouteDependencies{
		TaxonomyHandler:         deps.TaxonomyHandler,
		AuthMiddleware:          deps.AuthMiddleware,
		AccountStatusMiddleware: deps.AccountStatusMiddleware,
	})
	RegisterBusinessProfileRoutes(api, BusinessProfileRouteDependencies{
		BusinessProfileHandler:      deps.BusinessProfileHandler,
		BusinessProfileImageHandler: deps.BusinessProfileImageHandler,
		AuthMiddleware:              deps.AuthMiddleware,
		AccountStatusMiddleware:     deps.AccountStatusMiddleware,
	})
	RegisterNotificationPreferenceRoutes(api, NotificationPreferenceRouteDependencies{
		Handler:                 deps.NotificationPrefHandler,
		AuthMiddleware:          deps.AuthMiddleware,
		AccountStatusMiddleware: deps.AccountStatusMiddleware,
	})
	RegisterAccountPreferenceRoutes(api, AccountPreferenceRouteDependencies{
		Handler:                 deps.AccountPrefHandler,
		AuthMiddleware:          deps.AuthMiddleware,
		AccountStatusMiddleware: deps.AccountStatusMiddleware,
	})
}
