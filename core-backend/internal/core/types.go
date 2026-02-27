package core

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Cache    CacheConfig    `mapstructure:"cache"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Server   ServerConfig   `mapstructure:"server"`
}

type AppConfig struct {
	Name        string `mapstructure:"name"        validate:"required"`
	Environment string `mapstructure:"environment" validate:"required,oneof=development production testing local"`
	Version     string `mapstructure:"version"     validate:"required"`
	Port        int    `mapstructure:"port"        validate:"required,gt=0,lte=65535"`
	Debug       bool   `mapstructure:"debug"`
}

type DatabaseConfig struct {
	Type            string        `mapstructure:"type"              validate:"required,oneof=postgres postgresql mysql sqlserver"`
	Host            string        `mapstructure:"host"              validate:"required,hostname|ip"`
	Port            int           `mapstructure:"port"              validate:"required,gt=0,lte=65535"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"           validate:"required,oneof=disable require verify-full verify-ca"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"    validate:"gt=0"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"    validate:"gt=0"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" validate:"gt=0"`
}

type CacheConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Host     string `mapstructure:"host"      validate:"required_if=Enabled true,hostname|ip"`
	Port     int    `mapstructure:"port"      validate:"required_if=Enabled true,gt=0,lte=65535"`
	Password string `mapstructure:"password"  validate:"required_if=Enabled true"`
	DB       int    `mapstructure:"db"        validate:"gte=0"`
	PoolSize int    `mapstructure:"pool_size" validate:"required_if=Enabled true,gt=0"`
}

type LoggerConfig struct {
	Level            string   `mapstructure:"level"              validate:"required,oneof=debug info warn error"`
	Encoding         string   `mapstructure:"encoding"           validate:"required,oneof=json console"`
	OutputPaths      []string `mapstructure:"output_paths"`
	ErrorOutputPaths []string `mapstructure:"error_output_paths"`
}

type ServerConfig struct {
	ReadTimeout     time.Duration `mapstructure:"read_timeout"     validate:"gt=0"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"    validate:"gt=0"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"     validate:"gt=0"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" validate:"gt=0"`
}

type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
	WithContext(ctx context.Context) Logger
	Sync() error
}

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Health(ctx context.Context) error
	Close() error
}
