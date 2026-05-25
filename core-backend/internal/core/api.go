package core

import (
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	pkgmiddleware "github.com/Final-Year-Project-G22/backend/core/pkg/middleware"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
)

func NewHumaAPI(engine *gin.Engine, cfg *Config, log Logger) huma.API {
	errors.InitHumaErrorHandler()

	config := huma.DefaultConfig(cfg.App.Name, cfg.App.Version)
	config.Info.Description = "Backend API for the Adisu platform"

	api := humagin.New(engine, config)

	api.UseMiddleware(pkgmiddleware.LocaleResolver())

	log.Info("Huma API initialized",
		String("title", cfg.App.Name),
		String("version", cfg.App.Version),
	)

	return api
}
