package payment

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"go.uber.org/fx"
)

// Module is the payment module DI container.
// For Phase 2, it only registers entities with the SchemaManager so migrations can be generated.
// Full wiring (handlers, repositories, usecases) will be added in later phases.
var Module = fx.Module("payment",
	fx.Provide(NewEntityProvider),

	fx.Invoke(func(sm *core.SchemaManager, provider *EntityProvider) {
		if err := sm.RegisterProvider(provider); err != nil {
			panic(err)
		}
	}),
)
