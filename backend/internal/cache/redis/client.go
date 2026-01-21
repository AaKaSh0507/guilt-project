package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg Config) *redis.Client {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		// fallback to default host/port if parsing fails
		opts = &redis.Options{
			Addr: "localhost:6379",
		}
	}
	return redis.NewClient(opts)
}

func Ping(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}
