package main

import (
	"context"
	"flag"
	"os"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	appmodules "github.com/Final-Year-Project-G22/backend/core/internal/modules"
	"go.uber.org/fx"
)

func main() {
	action := flag.String("action", "", "Action: list, generate, apply, status")
	name := flag.String("name", "", "Migration name (for generate)")
	moduleFilter := flag.String("modules", "", "Comma-separated module names (empty = all)")
	flag.Parse()

	_ = os.Setenv("SKIP_AUTO_MIGRATIONS", "true")

	var migrator *core.Migrator
	var logger core.Logger

	app := fx.New(
		core.Module,
		appmodules.Modules,
		fx.Populate(&migrator, &logger),
		fx.NopLogger,
	)

	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		logger.Fatal(err.Error())
	}
	defer func() {
		if err := app.Stop(ctx); err != nil {
			logger.Error(err.Error())
		}
	}()

	switch *action {
	case "list":
		migrator.ListModules()
	case "generate":
		if *name == "" {
			logger.Fatal("Migration name is required for generate action ")
		}
		err := migrator.GenerateMigration(*name, *moduleFilter)
		if err != nil {
			logger.Fatal(err.Error())
			panic(err)
		}
	case "apply":
		err := migrator.ApplyMigrations()
		if err != nil {
			logger.Fatal(err.Error())
			panic(err)
		}
	case "status":
		migrator.CheckStatus()
	default:
		logger.Fatal("Invalid action. Use: list, generate, apply, or status")
	}
}
