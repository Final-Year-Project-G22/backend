package bootstrap

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"go.uber.org/fx"
)

type Application struct {
	*fx.App
}

func NewApp() *Application {
	app := fx.New(
		core.Module,
	)

	return &Application{app}
}
