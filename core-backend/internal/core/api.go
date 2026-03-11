package core

import (
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
)

func NewHumaAPI(engine *gin.Engine, cfg *Config, log Logger) huma.API {
	// Initialize custom error handler before creating API
	errors.InitHumaErrorHandler()

	config := huma.DefaultConfig(cfg.App.Name, cfg.App.Version)
	config.Info.Description = "Backend API for the Adisu platform"

	api := humagin.New(engine, config)

	log.Info("Huma API initialized",
		String("title", cfg.App.Name),
		String("version", cfg.App.Version),
	)

	return api
}
