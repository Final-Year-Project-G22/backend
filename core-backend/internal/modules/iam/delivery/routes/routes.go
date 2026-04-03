package routes

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/handler"
	"github.com/danielgtaylor/huma/v2"
)

type RouteDependencies struct {
	AuthHandler                *handler.AuthHandler
	AdminHandler               *handler.AdminHandler
	PermissionHandler          *handler.PermissionHandler
	RoleHandler                *handler.RoleHandler
	UserHandler                *handler.UserHandler
	ImageHandler               *handler.ImageHandler
	OAuthHandler               *handler.OAuthHandler
	AuthMiddleware             func(huma.Context, func(huma.Context))
	AccountStatusMiddleware    func(huma.Context, func(huma.Context))
	ReadPermissionMiddleware   func(huma.Context, func(huma.Context))
	WritePermissionMiddleware  func(huma.Context, func(huma.Context))
	UpdatePermissionMiddleware func(huma.Context, func(huma.Context))
	DeletePermissionMiddleware func(huma.Context, func(huma.Context))
}

func RegisterRoutes(api huma.API, deps RouteDependencies) {
	RegisterAuthRoutes(api, deps)
	RegisterUserRoutes(api, UserRouteDependencies{
		ImageHandler:            deps.ImageHandler,
		AuthMiddleware:          deps.AuthMiddleware,
		UserHandler:             deps.UserHandler,
		AccountStatusMiddleware: deps.AccountStatusMiddleware,
	})
	RegisterPermissionRoutes(api, PermissionRouteDependencies{
		PermissionHandler:        deps.PermissionHandler,
		AuthMiddleware:           deps.AuthMiddleware,
		AccountStatusMiddleware:  deps.AccountStatusMiddleware,
		ReadPermissionMiddleware: deps.ReadPermissionMiddleware,
	})
	RegisterRoleRoutes(api, RoleRouteDependencies{
		RoleHandler:                deps.RoleHandler,
		AuthMiddleware:             deps.AuthMiddleware,
		AccountStatusMiddleware:    deps.AccountStatusMiddleware,
		ReadPermissionMiddleware:   deps.ReadPermissionMiddleware,
		WritePermissionMiddleware:  deps.WritePermissionMiddleware,
		UpdatePermissionMiddleware: deps.UpdatePermissionMiddleware,
		DeletePermissionMiddleware: deps.DeletePermissionMiddleware,
	})
}
