package core

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(cfg *Config) (Logger, error) {
	var zapConfig zap.Config

	// Use preset configs based on the environment
	if cfg.IsDevelopment() || cfg.IsLocal() {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	// override with config values
	zapConfig.Level = zap.NewAtomicLevelAt(parseLevel(cfg.Logger.Level))
	zapConfig.Encoding = cfg.Logger.Encoding
	zapConfig.OutputPaths = cfg.Logger.OutputPaths
	zapConfig.ErrorOutputPaths = cfg.Logger.ErrorOutputPaths
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Build the logger
	zapLog, err := zapConfig.Build(
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return &zapLogger{logger: zapLog}, nil
}

type zapLogger struct {
	logger *zap.Logger
}

// Debug logs a debug message
func (l *zapLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

// Info logs an informational message
func (l *zapLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

// Warn logs a warning message
func (l *zapLogger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

// Error logs an error message
func (l *zapLogger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *zapLogger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

// With creates a child logger with additional fields
func (l *zapLogger) With(fields ...zap.Field) Logger {
	return &zapLogger{
		logger: l.logger.With(fields...),
	}
}

// WithContext creates a logger with context
func (l *zapLogger) WithContext(ctx context.Context) Logger {
	// Extract fields from context if needed
	// For now, just return the same logger
	return l
}

// Sync flushes any buffered log entries
func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

// parseLevel converts string level to zapcore.Level
func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// Convenience functions for zap fields
func String(key, val string) zap.Field      { return zap.String(key, val) }
func Int(key string, val int) zap.Field     { return zap.Int(key, val) }
func Int64(key string, val int64) zap.Field { return zap.Int64(key, val) }
func Bool(key string, val bool) zap.Field   { return zap.Bool(key, val) }
func Error(err error) zap.Field             { return zap.Error(err) }
func Any(key string, val any) zap.Field     { return zap.Any(key, val) }
