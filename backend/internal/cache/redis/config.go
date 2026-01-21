package redis

import (
	"fmt"
	"os"
)

type Config struct {
	URL string
}

func LoadConfig() (Config, error) {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		// fallback for local dev
		url = "redis://localhost:6379"
	}
	return Config{URL: url}, nil
}

func (c Config) String() string {
	return fmt.Sprintf("RedisConfig{URL=%s}", c.URL)
}
