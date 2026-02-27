package core

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database wraps gorm.DB
type Database struct {
	*gorm.DB
}

// The base model for all entities with tenant isolation
type BaseModel struct {
	ID        string         `gorm:"type:varchar(36);primaryKey"                json:"id"`
	TenantID  string         `gorm:"type:varchar(36);index:idx_tenant;not null" json:"tenant_id"`
	CreatedAt time.Time      `gorm:"autoCreateTime"                             json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"                             json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                      json:"-"`
}

// Before create hook to set ID and TenantID
func (b *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	// Set ID and TenantID logic here if needed
	return
}

func NewDatabase(cfg *Config, log Logger) (*Database, error) {
	log.Info("Connecting to database with GORM",
		String("host", cfg.Database.Host),
		Int("port", int(cfg.Database.Port)),
		String("database", cfg.Database.DBName),
	)

	// Configure gorm logger
	gormLogger := logger.New(
		&gormLoggerAdapter{log: log},
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  getGormLogLevel(cfg.Logger.Level),
			IgnoreRecordNotFoundError: true,
			Colorful:                  cfg.IsDevelopment(),
		},
	)

	// Open Database connection
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		Logger:                 gormLogger,
		SkipDefaultTransaction: true, // Improve performance
		PrepareStmt:            true, // Prepared statement cache,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(int(cfg.Database.MaxOpenConns))
	sqlDB.SetMaxIdleConns(int(cfg.Database.MaxIdleConns))
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Ping to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("Successfully connected to database with GORM")

	return &Database{DB: db}, nil
}

func (db *Database) Health(ctx context.Context) error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *Database) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// WithTenant returns a new DB instance scoped to a tenant
func (db *Database) WithTenant(tenantID string) *gorm.DB {
	return db.Where("tenant_id = ?", tenantID)
}

// Transaction executes a function within a database transaction
func (db *Database) Transaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return db.DB.WithContext(ctx).Transaction(fn)
}

type gormLoggerAdapter struct {
	log Logger
}

func (l *gormLoggerAdapter) Printf(format string, args ...interface{}) {
	l.log.Info(fmt.Sprintf(format, args...))
}

func getGormLogLevel(level string) logger.LogLevel {
	switch level {
	case "debug":
		return logger.Info
	case "info":
		return logger.Warn
	case "warn":
		return logger.Warn
	case "error":
		return logger.Error
	default:
		return logger.Warn
	}
}
