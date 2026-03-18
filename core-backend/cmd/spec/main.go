package main

import (
	"encoding/json"
	"os"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/handler"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/routes"
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

	routes.RegisterRoutes(api, routes.RouteDependencies{
		AuthHandler:    &handler.AuthHandler{},
		AuthMiddleware: noOpMiddleware,
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
