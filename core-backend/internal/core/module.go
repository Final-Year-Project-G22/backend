package core

import (
	"context"
	"net/http"
	"os"

	"github.com/Final-Year-Project-G22/backend/core/pkg/i18n"
	"github.com/Final-Year-Project-G22/backend/core/pkg/storage"
	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/fx"
)

var Module = fx.Module("core",
	fx.Provide(
		NewConfig,
		NewLogger,
		NewDatabase,
		NewSchemaManager,
		NewMigrator,
		NewCache,
		NewGinEngine,
		NewHTTPServer,
		NewHumaAPI,
		func(cfg *Config) storage.Config {
			return cfg.Storage
		},
		storage.NewStorage,
	),
	fx.Invoke(registerLifecycleHooks),
	// Ensure HTTP server and Huma API are instantiated by depending on them
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
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Core Module starting",
				String("cfg", cfg.App.Name),
				String("env", cfg.App.Environment),
			)
			// Run Health checks
			if err := db.Health(ctx); err != nil {
				return err
			}
			if os.Getenv("SKIP_AUTO_MIGRATIONS") == "true" {
				log.Info("Skipping automatic migrations")
				log.Info("Core Module started successfully")
				return nil
			}
			// Run Migrations
			log.Info("Running migrations")
			if err := m.ApplyMigrations(); err != nil {
				return err
			}

			log.Info("Initialize Internationalization")
			if err := i18n.Init(""); err != nil {
				return err
			}

			log.Info("Core Module started successfully")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Shutting down core module")
			var shutdownErr error

			// Close Connections
			if err := db.Close(); err != nil {
				log.Error("Failed to close database", Error(err))
				shutdownErr = err
			}

			if err := cache.Close(); err != nil {
				log.Error("Failed to close cache", Error(err))
				shutdownErr = err
			}

			// Sync logger last
			_ = log.Sync()

			return shutdownErr
		},
	})
}
