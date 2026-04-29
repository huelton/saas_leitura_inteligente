package database

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schemaSQL string

//go:embed schema_phase2.sql
var schemaPhase2SQL string

//go:embed schema_phase3.sql
var schemaPhase3SQL string

//go:embed schema_phase4.sql
var schemaPhase4SQL string

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	var n int
	err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'books'`).Scan(&n)
	if err != nil {
		return fmt.Errorf("migrate: check books table: %w", err)
	}
	if n == 0 {
		if _, err := pool.Exec(ctx, schemaSQL); err != nil {
			return fmt.Errorf("migrate: schema v1: %w", err)
		}
	}
	if _, err := pool.Exec(ctx, schemaPhase2SQL); err != nil {
		return fmt.Errorf("migrate: schema phase2: %w", err)
	}
	if _, err := pool.Exec(ctx, schemaPhase3SQL); err != nil {
		return fmt.Errorf("migrate: schema phase3: %w", err)
	}
	if _, err := pool.Exec(ctx, schemaPhase4SQL); err != nil {
		return fmt.Errorf("migrate: schema phase4: %w", err)
	}
	return nil
}

func EmbeddedSchema() string {
	return schemaSQL
}
