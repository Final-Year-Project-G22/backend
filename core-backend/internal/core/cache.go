package core

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	// redis.Nil indicates a cache miss
	return val, err
}

func (r *RedisCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisCache) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

func (r *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

func (r *RedisCache) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}

// NoOpCache is a cache implementation that does nothing
type NoOpCache struct{}

func (n *NoOpCache) Get(ctx context.Context, key string) (string, error) {
	return "", redis.Nil
}

func (n *NoOpCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return nil
}

func (n *NoOpCache) Delete(ctx context.Context, keys ...string) error {
	return nil
}

func (n *NoOpCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	return 0, nil
}

func (n *NoOpCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return nil
}

func (n *NoOpCache) Health(ctx context.Context) error {
	return nil
}

func (n *NoOpCache) Close() error {
	return nil
}

func NewCache(cfg *Config, log Logger) (Cache, error) {
	if !cfg.Cache.Enabled {
		log.Info("Cache is disabled, using NoOpCache")
		return &NoOpCache{}, nil
	}

	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Cache.Host, cfg.Cache.Port),
		Password: cfg.Cache.Password,
		DB:       cfg.Cache.DB,
		PoolSize: cfg.Cache.PoolSize,
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis cache", Error(err))
		return nil, fmt.Errorf("failed to connect to Redis cache: %w", err)
	}

	return &RedisCache{
		client: client,
	}, nil
}
