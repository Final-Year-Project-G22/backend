package routes

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/handler"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	"github.com/danielgtaylor/huma/v2"
)

type RouteDependencies struct {
	AuthHandler             *handler.AuthHandler
	AdminHandler            *handler.AdminHandler
	TaxonomyAdminHandler    *handler.TaxonomyAdminHandler
	PermissionHandler       *handler.PermissionHandler
	RoleHandler             *handler.RoleHandler
	UserHandler             *handler.UserHandler
	ImageHandler            *handler.ImageHandler
	OAuthHandler            *handler.OAuthHandler
	AuthMiddleware          func(huma.Context, func(huma.Context))
	AccountStatusMiddleware func(huma.Context, func(huma.Context))
	RoleAssignmentUsecase   usecase.RoleAssignmentUsecase
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
	})
}
