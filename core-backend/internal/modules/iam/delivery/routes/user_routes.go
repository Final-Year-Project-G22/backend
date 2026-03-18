package routes

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/handler"
	"github.com/danielgtaylor/huma/v2"
)

const (
	usersBase = "/api/v1/users"
)

type UserRouteDependencies struct {
	ImageHandler   *handler.ImageHandler
	AuthMiddleware func(huma.Context, func(huma.Context))
}

func RegisterUserRoutes(api huma.API, deps UserRouteDependencies) {
	huma.Register(api, huma.Operation{
		OperationID: "uploadAvatar",
		Method:      "POST",
		Path:        usersBase + "/avatar",
		Summary:     "Upload user avatar",
		Description: "Uploads and sets the authenticated user's avatar image",
		Tags:        []string{"Users"},
		Middlewares: huma.Middlewares{deps.AuthMiddleware},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, deps.ImageHandler.HandleUploadAvatar)
}
