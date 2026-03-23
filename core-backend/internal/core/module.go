package core

import (
	"context"
	"net/http"
	"os"

	"github.com/Final-Year-Project-G22/backend/core/internal/handlers"
	"github.com/Final-Year-Project-G22/backend/core/pkg/email"
	"github.com/Final-Year-Project-G22/backend/core/pkg/i18n"
	"github.com/Final-Year-Project-G22/backend/core/pkg/rabbitmq"
	"github.com/Final-Year-Project-G22/backend/core/pkg/storage"
	"github.com/danielgtaylor/huma/v2"

	"go.uber.org/fx"
)

func provideMessageBus(cfg *Config) (rabbitmq.Bus, error) {
	if !cfg.RabbitMQ.Enabled {
		return rabbitmq.NoOp(), nil
	}
	return rabbitmq.New(cfg.RabbitMQ)
}

func provideEmailer(cfg *Config) (email.Emailer, error) {
	if !cfg.Email.Enabled {
		return nil, nil
	}
	return email.NewEmailer(cfg.Email)
}

func registerEventHandlers(lc fx.Lifecycle, bus rabbitmq.Bus, emailer email.Emailer, logger Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return handlers.RegisterEventHandlers(bus, emailer, logger)
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}

var Module = fx.Module("core",
	fx.Provide(
		NewConfig,
		NewLogger,
		NewDatabase,
		NewSchemaManager,
		NewMigrator,
		NewCache,
		provideMessageBus,
		provideEmailer,
		NewGinEngine,
		NewHTTPServer,
		NewHumaAPI,
		func(cfg *Config) storage.Config {
			return cfg.Storage
		},
		storage.NewStorage,
	),
	fx.Invoke(registerLifecycleHooks),
	fx.Invoke(registerEventHandlers),
	fx.Invoke(func(*http.Server, huma.API) {}),
)

func registerLifecycleHooks(
	lc fx.Lifecycle,
	cfg *Config,
	log Logger,
	db *Database,
	sm *SchemaManager,
	m *Migrator,
	cache Cache,
	bus rabbitmq.Bus,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Core Module starting",
				String("cfg", cfg.App.Name),
				String("env", cfg.App.Environment),
			)
			log.Info("Initialize Internationalization")
			if err := i18n.Init(""); err != nil {
				return err
			}
			if err := db.Health(ctx); err != nil {
				return err
			}
			if cfg.RabbitMQ.Enabled {
				log.Info("Connecting to RabbitMQ",
					String("host", cfg.RabbitMQ.Host),
					Int("port", cfg.RabbitMQ.Port),
				)
			}
			if cfg.Email.Enabled {
				log.Info("Email service configured",
					String("host", cfg.Email.Host),
					Int("port", cfg.Email.Port),
				)
			}
			if os.Getenv("SKIP_AUTO_MIGRATIONS") == "true" {
				log.Info("Skipping automatic migrations")
				log.Info("Core Module started successfully")
				return nil
			}
			log.Info("Running migrations")
			if err := m.ApplyMigrations(); err != nil {
				return err
			}

			log.Info("Core Module started successfully")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Shutting down core module")
			var shutdownErr error

			if err := db.Close(); err != nil {
				log.Error("Failed to close database", Error(err))
				shutdownErr = err
			}

			if err := cache.Close(); err != nil {
				log.Error("Failed to close cache", Error(err))
				shutdownErr = err
			}

			if cfg.RabbitMQ.Enabled && bus != nil {
				if err := bus.Close(); err != nil {
					log.Error("Failed to close RabbitMQ", Error(err))
					shutdownErr = err
				}
			}

			_ = log.Sync()

			return shutdownErr
		},
	})
}
