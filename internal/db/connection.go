package db

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

type DB = *sql.DB

func NewDB(ctx context.Context, url string) (DB, error) {
	// Parse PGX config
	cfg, err := pgx.ParseConfig(url)
	if err != nil {
		return nil, err
	}

	// Register config as stdlib DSN
	dsn := stdlib.RegisterConnConfig(cfg)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// Optional health check
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

func MustDB(ctx context.Context, url string) DB {
	db, err := NewDB(ctx, url)
	if err != nil {
		panic(err)
	}
	return db
}
