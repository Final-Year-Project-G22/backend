package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/port"
	"github.com/redis/go-redis/v9"
)

type ConversationCache struct {
	cache core.Cache
}

func NewConversationCache(cache core.Cache) port.ConversationCachePort {
	return &ConversationCache{cache: cache}
}

func (c *ConversationCache) Get(ctx context.Context, key string, out any) (bool, error) {
	value, err := c.cache.Get(ctx, key)
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal([]byte(value), out); err != nil {
		return false, err
	}
	return true, nil
}

func (c *ConversationCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.cache.Set(ctx, key, string(b), ttl)
}

func (c *ConversationCache) Invalidate(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.cache.Delete(ctx, keys...)
}
