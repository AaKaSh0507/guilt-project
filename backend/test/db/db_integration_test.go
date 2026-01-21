package db_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestDBIntegration(t *testing.T) {
	url := os.Getenv("TEST_DB_URL")
	if url == "" {
		t.Fatal("TEST_DB_URL not set")
	}

	ctx := context.Background()
	db, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close(ctx)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("db ping failed: %v", err)
	}

	// sanity: check a table from migrations exists
	var count int
	if err := db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		t.Fatalf("failed to query users: %v", err)
	}

	// success just means migrations applied
}
