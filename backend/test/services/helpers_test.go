package services_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	url := os.Getenv("TEST_DB_URL")
	if url == "" {
		t.Fatalf("TEST_DB_URL not set")
	}

	cfg, err := pgx.ParseConfig(url)
	if err != nil {
		t.Fatalf("failed to parse db config: %v", err)
	}

	dsn := stdlib.RegisterConnConfig(cfg)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db failed: %v", err)
	}

	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("ping db failed: %v", err)
	}

	return db
}

func openTestRedis(t *testing.T) string {
	t.Helper()
	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		t.Fatalf("TEST_REDIS_URL not set")
	}
	return url
}
