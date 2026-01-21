package cache_test

import (
	"context"
	"os"
	"testing"
	"time"

	cache "guiltmachine/internal/cache"
	cacheRedis "guiltmachine/internal/cache/redis"
)

func TestRedisIntegration(t *testing.T) {
	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		t.Fatal("TEST_REDIS_URL not set")
	}

	cfg := cacheRedis.Config{URL: url}
	client := cacheRedis.NewRedisClient(cfg)

	if err := cacheRedis.Ping(context.Background(), client); err != nil {
		t.Fatalf("failed ping redis: %v", err)
	}

	r := cacheRedis.NewRedisCache(client)
	key := "integration:test:key"
	val := []byte("hello")

	if err := r.Set(context.Background(), key, val, 2*time.Second); err != nil {
		t.Fatalf("redis set failed: %v", err)
	}

	b, err := r.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("redis get failed: %v", err)
	}

	if string(b) != "hello" {
		t.Fatalf("unexpected value: %s", string(b))
	}

	if err := r.Del(context.Background(), key); err != nil {
		t.Fatalf("redis del failed: %v", err)
	}

	_, err = r.Get(context.Background(), key)
	if err == nil {
		t.Fatalf("expected not found after delete")
	}
	if err != cache.ErrNotFound {
		t.Fatalf("unexpected error: %v", err)
	}
}
