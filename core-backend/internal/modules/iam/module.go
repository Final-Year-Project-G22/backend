package iam

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"go.uber.org/fx"
)

var Module = fx.Module("iam",
	// Provide entity provider
	fx.Provide(
		fx.Annotate(
			NewEntityProvider,
			fx.As(new(core.EntityProvider)),
		),
	),

	// Auto-register with schema manager
	fx.Invoke(func(sm *core.SchemaManager, provider core.EntityProvider) {
		if err := sm.RegisterProvider(provider); err != nil {
			panic(err)
		}
	}),
)
