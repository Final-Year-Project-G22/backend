package infrastructure

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/port"
	"github.com/redis/go-redis/v9"
)

const (
	IngestToggleKey = "ingestion:toggle"
	IngestToggleTTL = 24 * time.Hour
)

type IngestToggle struct {
	cache  core.Cache
	logger core.Logger
	mu     sync.RWMutex
}

func NewIngestToggle(cache core.Cache, logger core.Logger) *IngestToggle {
	return &IngestToggle{
		cache:  cache,
		logger: logger,
	}
}

var _ port.IngestControl = (*IngestToggle)(nil)

func (t *IngestToggle) IsEnabled(ctx context.Context) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	val, err := t.cache.Get(ctx, IngestToggleKey)
	if errors.Is(err, redis.Nil) {
		return true
	}
	if err != nil {
		t.logger.Warn("Failed to get ingest toggle state", core.Error(err))
		return true
	}

	return val == "true"
}

func (t *IngestToggle) SetEnabled(ctx context.Context, enabled bool) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	value := "false"
	if enabled {
		value = "true"
	}

	err := t.cache.Set(ctx, IngestToggleKey, value, IngestToggleTTL)
	if err != nil {
		t.logger.Error("Failed to set ingest toggle", core.Error(err))
		return err
	}

	t.logger.Info("Ingest toggle updated", core.Bool("enabled", enabled))
	return nil
}

func (t *IngestToggle) GetToggleState(ctx context.Context) (bool, bool, error) {
	val, err := t.cache.Get(ctx, IngestToggleKey)
	if errors.Is(err, redis.Nil) {
		return true, false, nil
	}
	if err != nil {
		return false, false, err
	}

	return val == "true", true, nil
}
