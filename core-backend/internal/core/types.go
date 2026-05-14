package core

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/pkg/email"
	"github.com/Final-Year-Project-G22/backend/core/pkg/rabbitmq"
	"github.com/Final-Year-Project-G22/backend/core/pkg/storage"
	"go.uber.org/zap"
)

type Config struct {
	App       AppConfig       `mapstructure:"app"`
	AI        AIConfig        `mapstructure:"ai"`
	Ingestion IngestionConfig `mapstructure:"ingestion"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Cache     CacheConfig     `mapstructure:"cache"`
	Logger    LoggerConfig    `mapstructure:"logger"`
	Server    ServerConfig    `mapstructure:"server"`
	Storage   storage.Config  `mapstructure:"storage"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	OAuth     OAuthConfig     `mapstructure:"oauth"`
	RabbitMQ  rabbitmq.Config `mapstructure:"rabbitmq"`
	Email     email.Config    `mapstructure:"email"`
	Resend    ResendConfig    `mapstructure:"resend"`
	Chapa     ChapaConfig     `mapstructure:"chapa"`
	FCM       FCMConfig       `mapstructure:"fcm"`
}

type ResendConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	APIKey        string `mapstructure:"api_key"`
	WebhookSecret string `mapstructure:"webhook_secret"`
	FromEmail     string `mapstructure:"from_email"`
	FromName      string `mapstructure:"from_name"`
}

type FCMConfig struct {
	CredentialsFile string `mapstructure:"credentials_file"`
}

type ChapaConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	SecretKey     string `mapstructure:"secret_key"`
	PublicKey     string `mapstructure:"public_key"`
	WebhookSecret string `mapstructure:"webhook_secret"`
	BaseURL       string `mapstructure:"base_url"`
	CallbackURL   string `mapstructure:"callback_url"`
	ReturnURL     string `mapstructure:"return_url"`
}

type AppConfig struct {
	Name        string `mapstructure:"name"        validate:"required"`
	Environment string `mapstructure:"environment" validate:"required,oneof=development production testing local"`
	Version     string `mapstructure:"version"     validate:"required"`
	Port        int    `mapstructure:"port"        validate:"required,gt=0,lte=65535"`
	GRPCPort    int    `mapstructure:"grpc_port"   validate:"required,gt=0,lte=65535"`
	Debug       bool   `mapstructure:"debug"`
}

type AIConfig struct {
	InferenceGRPCEndpoint string        `mapstructure:"inference_grpc_endpoint"`
	InferenceAuthToken    string        `mapstructure:"inference_auth_token"`
	InferenceTimeout      time.Duration `mapstructure:"inference_timeout" validate:"gt=0"`
	AskEnabled            bool          `mapstructure:"ask_enabled"`
	ConversationCacheTTL  time.Duration `mapstructure:"conversation_cache_ttl"`
}

type IngestionConfig struct {
	Enabled    bool                      `mapstructure:"enabled"`
	Signing    IngestionSigningConfig    `mapstructure:"signing"`
	Dispatcher IngestionDispatcherConfig `mapstructure:"dispatcher"`
}

type IngestionDispatcherConfig struct {
	BatchSize            int           `mapstructure:"batch_size"`
	Interval             time.Duration `mapstructure:"interval"`
	RetryBaseDelay       time.Duration `mapstructure:"retry_base_delay"`
	RetryMaxDelay        time.Duration `mapstructure:"retry_max_delay"`
	MaxAttemptsBeforeDLQ int           `mapstructure:"max_attempts_before_dlq"`
}

type IngestionSigningConfig struct {
	ActiveKeyID     string                `mapstructure:"active_key_id"`
	ActiveKeySecret string                `mapstructure:"active_key_secret"`
	PreviousKeys    []IngestionSigningKey `mapstructure:"previous_keys"`
}

type IngestionSigningKey struct {
	KeyID  string `mapstructure:"key_id"`
	Secret string `mapstructure:"secret"`
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

type JWTConfig struct {
	Secret          string        `mapstructure:"secret"            validate:"required,min=32"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"  validate:"required,gt=0"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl" validate:"required,gt=0"`
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

type OAuthConfig struct {
	EncryptionKey         string `mapstructure:"encryption_key" validate:"required,min=32"`
	CookieDomain          string `mapstructure:"cookie_domain"  validate:"required"`
	MobileRedirectBaseURL string `mapstructure:"mobile_redirect_base_url"`
	Providers             []OAuthProviderConfig
}

type OAuthProviderConfig struct {
	Name         string   `mapstructure:"name"         validate:"required"`
	ClientID     string   `mapstructure:"client_id"     validate:"required"`
	ClientSecret string   `mapstructure:"client_secret" validate:"required"`
	RedirectURI  string   `mapstructure:"redirect_uri"  validate:"required"`
	Scopes       []string `mapstructure:"scopes"        validate:"required"`
}
