package cache

import (
	"context"
	"time"
)

type Cache interface {
	// Basic KV
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	Del(ctx context.Context, key string) error

	// Counter
	Incr(ctx context.Context, key string, ttl time.Duration) (int64, error)

	// Lists
	Push(ctx context.Context, key string, value []byte, ttl time.Duration) (int64, error)
	Range(ctx context.Context, key string) ([][]byte, error)

	// Sets
	Sadd(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Scard(ctx context.Context, key string) (int64, error)
}
