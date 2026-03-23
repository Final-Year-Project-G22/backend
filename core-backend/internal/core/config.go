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

	// Bind JWT secret from environment variable
	err = v.BindEnv("jwt.secret", "JWT_SECRET")
	if err != nil {
		return nil, fmt.Errorf("error binding env jwt.secret: %w", err)
	}

	// Bind RabbitMQ config from environment variables
	err = v.BindEnv("rabbitmq.enabled", "RABBITMQ_ENABLED")
	if err != nil {
		return nil, fmt.Errorf("error binding env rabbitmq.enabled: %w", err)
	}
	err = v.BindEnv("rabbitmq.host", "RABBITMQ_HOST")
	if err != nil {
		return nil, fmt.Errorf("error binding env rabbitmq.host: %w", err)
	}
	err = v.BindEnv("rabbitmq.port", "RABBITMQ_PORT")
	if err != nil {
		return nil, fmt.Errorf("error binding env rabbitmq.port: %w", err)
	}
	err = v.BindEnv("rabbitmq.username", "RABBITMQ_USERNAME")
	if err != nil {
		return nil, fmt.Errorf("error binding env rabbitmq.username: %w", err)
	}
	err = v.BindEnv("rabbitmq.password", "RABBITMQ_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("error binding env rabbitmq.password: %w", err)
	}
	err = v.BindEnv("rabbitmq.vhost", "RABBITMQ_VHOST")
	if err != nil {
		return nil, fmt.Errorf("error binding env rabbitmq.vhost: %w", err)
	}
	err = v.BindEnv("rabbitmq.exchange", "RABBITMQ_EXCHANGE")
	if err != nil {
		return nil, fmt.Errorf("error binding env rabbitmq.exchange: %w", err)
	}
	err = v.BindEnv("rabbitmq.service_name", "RABBITMQ_SERVICE_NAME")
	if err != nil {
		return nil, fmt.Errorf("error binding env rabbitmq.service_name: %w", err)
	}

	// Bind SMTP config from environment variables
	err = v.BindEnv("email.enabled", "SMTP_ENABLED")
	if err != nil {
		return nil, fmt.Errorf("error binding env email.enabled: %w", err)
	}
	err = v.BindEnv("email.host", "SMTP_HOST")
	if err != nil {
		return nil, fmt.Errorf("error binding env email.host: %w", err)
	}
	err = v.BindEnv("email.port", "SMTP_PORT")
	if err != nil {
		return nil, fmt.Errorf("error binding env email.port: %w", err)
	}
	err = v.BindEnv("email.username", "SMTP_USERNAME")
	if err != nil {
		return nil, fmt.Errorf("error binding env email.username: %w", err)
	}
	err = v.BindEnv("email.password", "SMTP_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("error binding env email.password: %w", err)
	}
	err = v.BindEnv("email.from", "SMTP_FROM")
	if err != nil {
		return nil, fmt.Errorf("error binding env email.from: %w", err)
	}
	err = v.BindEnv("email.from_name", "SMTP_FROM_NAME")
	if err != nil {
		return nil, fmt.Errorf("error binding env email.from_name: %w", err)
	}
	err = v.BindEnv("email.enable_tls", "SMTP_ENABLE_TLS")
	if err != nil {
		return nil, fmt.Errorf("error binding env email.enable_tls: %w", err)
	}

	// 4. Set defaults
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.port", 4000)
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

	// 8. Resolve ${VAR} / ${VAR:-default} placeholders from environment
	if err := resolveConfigPlaceholders(&cfg, false); err != nil {
		return nil, fmt.Errorf("error resolving config placeholders: %w", err)
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

	// Conditional RabbitMQ validation
	if c.RabbitMQ.Enabled {
		rabbitValidate := validator.New()
		if err := rabbitValidate.Struct(c.RabbitMQ); err != nil {
			return fmt.Errorf("rabbitmq validation failed: %w", err)
		}
	}

	// Conditional email validation
	if c.Email.Enabled {
		emailValidate := validator.New()
		if err := emailValidate.Struct(c.Email); err != nil {
			return fmt.Errorf("email validation failed: %w", err)
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
