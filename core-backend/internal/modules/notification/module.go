package notification

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/application/service"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"notification",
	fx.Provide(NewEntityProvider),
	fx.Provide(service.NewDeliveryWorker),
	fx.Invoke(func(sm *core.SchemaManager, provider *EntityProvider) {
		if err := sm.RegisterProvider(provider); err != nil {
			panic(err)
		}
	}),
	fx.Invoke(func(lc fx.Lifecycle, worker *service.DeliveryWorker) {
		ctx, cancel := context.WithCancel(context.Background())
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				worker.Start(ctx)
				return nil
			},
			OnStop: func(context.Context) error {
				cancel()
				return nil
			},
		})
	}),
)
