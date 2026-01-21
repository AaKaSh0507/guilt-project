package repo_test

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
		t.Fatalf("TEST_DB_URL not set (expected from test/test_main.go)")
	}

	cfg, err := pgx.ParseConfig(url)
	if err != nil {
		t.Fatalf("failed to parse db config: %v", err)
	}

	dsn := stdlib.RegisterConnConfig(cfg)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("failed to ping test db: %v", err)
	}

	return db
}
