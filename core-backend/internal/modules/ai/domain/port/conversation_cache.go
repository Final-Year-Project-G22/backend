package port

import (
	"context"
	"time"
)

type ConversationCachePort interface {
	Get(ctx context.Context, key string, out any) (bool, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Invalidate(ctx context.Context, keys ...string) error
}
