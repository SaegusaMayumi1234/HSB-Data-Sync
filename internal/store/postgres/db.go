package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConfig struct {
	DSN string
}

type Database struct {
	Pool *pgxpool.Pool
}

// New creates a new PostgreSQL connection pool.
func New(ctx context.Context, cfg PostgresConfig) (*Database, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("pgxpool new: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pgxpool ping: %w", err)
	}

	return &Database{Pool: pool}, nil
}

// Close closes the connection pool.
func (db *Database) Close() {
	db.Pool.Close()
}
