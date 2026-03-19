package routes

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/handler"
	"github.com/danielgtaylor/huma/v2"
)

type RouteDependencies struct {
	AuthHandler    *handler.AuthHandler
	UserHandler    *handler.UserHandler
	ImageHandler   *handler.ImageHandler
	AuthMiddleware func(huma.Context, func(huma.Context))
}

func RegisterRoutes(api huma.API, deps RouteDependencies) {
	RegisterAuthRoutes(api, deps)
	RegisterUserRoutes(api, UserRouteDependencies{
		ImageHandler:   deps.ImageHandler,
		AuthMiddleware: deps.AuthMiddleware,
		UserHandler:    deps.UserHandler,
	})
}
