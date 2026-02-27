package core

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/pkg/utils"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

const (
	DBTypePostgres   = "postgres"
	DBTypePostgresql = "postgresql"
	DBTypeMySQL      = "mysql"
)

func NewConfig() (*Config, error) {
	// 1. Load .env file into actual environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found (this is optional): %v", err)
	}

	v := viper.New()

	// 2. Set up environment variable mapping FIRST
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 3. Explicitly bind environment variables to the correct nested keys
	// This is crucial for proper unmarshaling
	err := v.BindEnv("database.user", "DATABASE_USER")
	if err != nil {
		return nil, fmt.Errorf("error binding env database.user %w", err)
	}
	err = v.BindEnv("database.password", "DATABASE_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("error binding env database.password %w", err)
	}
	err = v.BindEnv("database.dbname", "DATABASE_DBNAME")
	if err != nil {
		return nil, fmt.Errorf("error binding env database.name %w", err)
	}

	// 4. Set defaults
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.port", 8080)
	v.SetDefault("logger.level", "info")
	v.SetDefault("database.sslmode", "disable")

	// 5. Read Base Config File (config.yaml)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("internal/configs")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// 6. Read environment specific config file
	env := v.GetString("app.environment")
	if env != "" {
		v.SetConfigName(fmt.Sprintf("config.%s", env))
		v.AddConfigPath("internal/configs")
		v.AddConfigPath(".")
		if err := v.MergeInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("error reading %s config file: %w", env, err)
			}
		}
	}

	// 7. Unmarshal the final config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	validate := validator.New()
	err := validate.StructExcept(c, "Cache", "Database.User", "Database.Password", "Database.DBName")

	if err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// Conditional cache validation
	if c.Cache.Enabled {
		cacheValidate := validator.New()
		if err := cacheValidate.Struct(c.Cache); err != nil {
			return fmt.Errorf("cache validation failed: %w", err)
		}
	}

	// Validate App.Environment (used directly in exec.Command)
	if err := utils.IsSafeString(c.App.Environment); err != nil {
		return fmt.Errorf("invalid app.environment: %w", err)
	}

	// Validate DSN components (used in GetDSN, which is then used in exec.Command)
	if err := utils.IsSafeDSNComponent(c.Database.Host); err != nil {
		return fmt.Errorf("invalid database.host: %w", err)
	}
	if err := utils.IsSafeString(c.Database.DBName); err != nil {
		return fmt.Errorf("invalid database.dbname: %w", err)
	}
	if err := utils.IsSafeString(c.Database.SSLMode); err != nil {
		return fmt.Errorf("invalid database.sslmode: %w", err)
	}

	return nil
}

func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

func (c *Config) IsLocal() bool {
	return c.App.Environment == "local"
}

func (c *Config) GetDatabaseUrl() string {
	switch c.Database.Type {
	case DBTypePostgres, DBTypePostgresql:
		return strings.TrimSpace(c.getPostgreSQLURL())
	case DBTypeMySQL:
		return c.getMySQLURL()
	default:
		return c.getPostgreSQLURL()
	}
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

func (c *Config) GetAddr() string {
	if !c.Cache.Enabled {
		return ""
	}

	return fmt.Sprintf("%s:%d", c.Cache.Host, c.Cache.Port)
}

func (c *Config) getPostgreSQLURL() string {
	encodedPassword := url.QueryEscape(c.Database.Password)
	encodedUser := url.QueryEscape(c.Database.User)

	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		encodedUser, encodedPassword, c.Database.Host,
		c.Database.Port, c.Database.DBName, c.Database.SSLMode)
}

func (c *Config) getMySQLURL() string {
	encodedPassword := url.QueryEscape(c.Database.Password)
	return fmt.Sprintf("mysql://%s:%s@%s:%d/%s",
		c.Database.User, encodedPassword, c.Database.Host,
		c.Database.Port, c.Database.DBName)
}

// GetCacheURL provides a standard URL format for cache connection
func (c *Config) GetCacheURL() string {
	if !c.Cache.Enabled {
		return ""
	}

	if c.Cache.Password == "" {
		return fmt.Sprintf("redis://%s:%d", c.Cache.Host, c.Cache.Port)
	}
	return fmt.Sprintf("redis://:%s@%s:%d", c.Cache.Password, c.Cache.Host, c.Cache.Port)
}

func (c *Config) IsCacheEnabled() bool {
	return c.Cache.Enabled
}
