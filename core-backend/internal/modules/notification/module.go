package notification

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"notification",
	fx.Provide(NewEntityProvider),
	fx.Invoke(func(sm *core.SchemaManager, provider *EntityProvider) {
		if err := sm.RegisterProvider(provider); err != nil {
			panic(err)
		}
	}),
)
