package routes

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/delivery/handler"
	"github.com/danielgtaylor/huma/v2"
)

type RouteDependencies struct {
	GuideViewHandler           *handler.GuideViewHandler
	GuideAdminHandler          *handler.GuideAdminHandler
	AuthMiddleware             func(huma.Context, func(huma.Context))
	AccountStatusMiddleware    func(huma.Context, func(huma.Context))
	ReadPermissionMiddleware   func(huma.Context, func(huma.Context))
	WritePermissionMiddleware  func(huma.Context, func(huma.Context))
	UpdatePermissionMiddleware func(huma.Context, func(huma.Context))
	DeletePermissionMiddleware func(huma.Context, func(huma.Context))
}

func RegisterRoutes(api huma.API, deps RouteDependencies) {
	RegisterGuideViewRoutes(api, deps)
	RegisterGuideAdminRoutes(api, deps)
}
