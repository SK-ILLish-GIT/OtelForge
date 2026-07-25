package db

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/001_initial.sql
var migration001 string

//go:embed migrations/002_ssh_private_key.sql
var migration002 string

//go:embed migrations/003_ssh_host_key_fingerprint.sql
var migration003 string

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	for i, sql := range []string{migration001, migration002, migration003} {
		if _, err := pool.Exec(ctx, sql); err != nil {
			return fmt.Errorf("run migration %d: %w", i+1, err)
		}
	}
	return nil
}
