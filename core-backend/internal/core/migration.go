package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Migrator struct {
	schema *SchemaManager
	config *Config
	log    Logger
	db     *Database
}

func NewMigrator(
	schema *SchemaManager,
	config *Config,
	logger Logger,
	db *Database,
) *Migrator {
	return &Migrator{
		schema: schema,
		config: config,
		log:    logger,
		db:     db,
	}
}

func (m *Migrator) ListModules() {
	m.log.Info("Listing registered modules")

	m.log.Info("\n Registered Modules:")

	totalEntities := 0
	for module, count := range m.schema.GetModuleInfo() {
		fmt.Printf("  %-15s %d entities\n", module, count)
		m.log.Debug("Module registered",
			String("module", module),
			Int("entities", count))
		totalEntities += count
	}

	m.log.Info("entities", Int("total_entities", totalEntities))
}

func (m *Migrator) GenerateMigration(migrationName, moduleFilter string) error {
	env := m.config.App.Environment

	var entities []interface{}
	var targetModules string

	if moduleFilter == "" {
		entities = m.schema.GetAllEntities()
		targetModules = "ALL"
		m.log.Info("Generating migration for all module",
			Int("total_entities", len(entities)))
	} else {
		mods := strings.Split(moduleFilter, ",")
		for i, mod := range mods {
			mods[i] = strings.TrimSpace(mod)
		}
		entities = m.schema.GetEntitiesByModules(mods...)
		targetModules = strings.Join(mods, ",")
		m.log.Info("Generating migration for specific modules",
			Int("total_entities", len(entities)))
	}

	m.log.Info("Generating Migration")
	m.log.Info("Name: ", String("name", migrationName))
	m.log.Info("Modules: ", String("modules", targetModules))
	m.log.Info("Entities: ", Int("entities", len(entities)))
	m.log.Info("Env: ", String("env", env))

	// Schema file
	schemaFile := "schema/schema.sql"

	file, err := os.Create(schemaFile)
	if err != nil {
		m.log.Fatal("Failed to create schema file", Error(err))
		return err
	}

	// Load gorm entities
	err = m.schema.LoadGORMSchema(file, m.config, m.db)
	if err != nil {
		m.log.Fatal("Failed to load gorm schema", Error(err))
		return err
	}
	err = file.Close()
	if err != nil {
		m.log.Fatal("Failed to close schema file", Error(err))
		return err
	}

	m.log.Info("Schema file created", String("schema_file", schemaFile))
	m.log.Info("Migration name", String("migration_name", migrationName))

	// nolint:gosec // G204: Arguments are derived from validated application configuration, not untrusted user input.
	cmd := exec.Command("atlas", "migrate", "diff",
		migrationName,
		"--to", fmt.Sprintf("file://%s", schemaFile),
		"--dir", "file://migrations",
		"--dev-url", "docker://postgres?search_path=public",
	)

	m.log.Debug("Atlas command", String("command", cmd.String()))

	m.log.Info("Running Atlas migration generation",
		String("migration_name", migrationName),
		String("environment", env))

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		m.log.Fatal("Migration generation failed", Error(err))
		return err
	}

	m.log.Info("Migration generated successfully", String("migrationName", migrationName))
	fmt.Println("\n Migration generated successfully!")

	return nil
}

func (m *Migrator) CheckStatus() {
	m.log.Info("Migration Status")
	m.log.Info("Environment: %s", String("env", m.config.App.Environment))

	// nolint:gosec // G204: Arguments are derived from validated application configuration, not untrusted user input.
	cmd := exec.Command("atlas", "migrate", "status",
		"--dir", "file://migrations",
		"--env", m.config.App.Environment,
		"--dev-url", "docker://postgres:latest",
		"--to", m.config.GetDatabaseUrl(),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		m.log.Error("Failed to check migration status", Error(err))
		m.log.Fatal("Status check failed")
	}
}

func (m *Migrator) ApplyMigrations() error {
	m.log.Info("Starting migration process...")
	m.log.Info("Migration directory: file://migrations")
	m.log.Info("Target database", String("url", m.config.GetDatabaseUrl()))

	// nolint:gosec // G204: Arguments are derived from validated application configuration, not untrusted user input.
	cmd := exec.Command("atlas", "migrate", "apply",
		"--dir", "file://migrations",
		"--url", m.config.GetDatabaseUrl(),
	)

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Log each line of Atlas output to see migration details
	if outputStr != "" {
		m.log.Info("Atlas migration output:")
		for _, line := range strings.Split(strings.TrimSpace(outputStr), "\n") {
			if line != "" {
				m.log.Info("  " + line)
			}
		}
	}

	if err != nil {
		m.log.Error("Migration failed", Error(err))
		return fmt.Errorf("atlas failed: %w, output: %s", err, outputStr)
	}

	m.log.Info("Migrations applied successfully!")
	return nil
}

func (m *Migrator) Entities() []interface{} {
	return m.schema.GetAllEntities()
}
