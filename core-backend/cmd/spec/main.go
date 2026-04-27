package main

import (
	"encoding/json"
	"os"

	communityhandler "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/delivery/handler"
	communityroutes "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/delivery/routes"
	iamhandler "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/handler"
	iamroutes "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/routes"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.New()

	config := huma.DefaultConfig("Adisu Backend API", "1.0.0")
	config.Info.Description = "Backend API for the Adisu platform"

	api := humagin.New(engine, config)

	noOpMiddleware := func(ctx huma.Context, next func(huma.Context)) {
		next(ctx)
	}

	iamroutes.RegisterRoutes(api, iamroutes.RouteDependencies{
		AuthHandler:                &iamhandler.AuthHandler{},
		AdminHandler:               &iamhandler.AdminHandler{},
		PermissionHandler:          &iamhandler.PermissionHandler{},
		RoleHandler:                &iamhandler.RoleHandler{},
		UserHandler:                &iamhandler.UserHandler{},
		ImageHandler:               &iamhandler.ImageHandler{},
		OAuthHandler:               &iamhandler.OAuthHandler{},
		AuthMiddleware:             noOpMiddleware,
		AccountStatusMiddleware:    noOpMiddleware,
		ReadPermissionMiddleware:   noOpMiddleware,
		WritePermissionMiddleware:  noOpMiddleware,
		UpdatePermissionMiddleware: noOpMiddleware,
		DeletePermissionMiddleware: noOpMiddleware,
	})

	communityroutes.RegisterRoutes(api, communityroutes.RouteDependencies{
		CommunityHandler:           &communityhandler.CommunityHandler{},
		CommunityAdminHandler:      &communityhandler.CommunityAdminHandler{},
		AuthMiddleware:             noOpMiddleware,
		AccountStatusMiddleware:    noOpMiddleware,
		ReadPermissionMiddleware:   noOpMiddleware,
		WritePermissionMiddleware:  noOpMiddleware,
		UpdatePermissionMiddleware: noOpMiddleware,
		DeletePermissionMiddleware: noOpMiddleware,
	})

	file, err := os.Create("docs/openapi.json")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(api.OpenAPI()); err != nil {
		panic(err)
	}
}
