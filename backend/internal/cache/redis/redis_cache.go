package redis

import (
	"context"
	"errors"
	"time"

	libredis "github.com/redis/go-redis/v9"

	"guiltmachine/internal/cache"
)

type RedisCache struct {
	client *libredis.Client
}

func NewRedisCache(client *libredis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (r *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl > 0 {
		return r.client.Set(ctx, key, value, ttl).Err()
	}
	return r.client.Set(ctx, key, value, 0).Err()
}

func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Bytes()
	if errors.Is(err, libredis.Nil) {
		return nil, cache.ErrNotFound
	}
	return val, err
}

func (r *RedisCache) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	val, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if ttl > 0 {
		_ = r.client.Expire(ctx, key, ttl).Err()
	}
	return val, nil
}

func (r *RedisCache) Push(ctx context.Context, key string, value []byte, ttl time.Duration) (int64, error) {
	l, err := r.client.RPush(ctx, key, value).Result()
	if err != nil {
		return 0, err
	}
	if ttl > 0 {
		_ = r.client.Expire(ctx, key, ttl).Err()
	}
	return l, nil
}

func (r *RedisCache) Range(ctx context.Context, key string) ([][]byte, error) {
	items, err := r.client.LRange(ctx, key, 0, -1).Result()
	if errors.Is(err, libredis.Nil) {
		return nil, cache.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	out := make([][]byte, len(items))
	for i, v := range items {
		out[i] = []byte(v)
	}
	return out, nil
}

func (r *RedisCache) Sadd(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	err := r.client.SAdd(ctx, key, value).Err()
	if err != nil {
		return err
	}
	if ttl > 0 {
		_ = r.client.Expire(ctx, key, ttl).Err()
	}
	return nil
}

func (r *RedisCache) Scard(ctx context.Context, key string) (int64, error) {
	val, err := r.client.SCard(ctx, key).Result()
	if errors.Is(err, libredis.Nil) {
		return 0, cache.ErrNotFound
	}
	return val, err
}
