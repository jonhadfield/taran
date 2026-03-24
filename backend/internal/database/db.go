package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	// Pool tuning for Cloud Run (concurrency=80, auto-scaling instances).
	// Neon serverless PostgreSQL wakes from cold in ~500ms, so set a
	// reasonable connect timeout and keep idle connections low.
	config.MaxConns = 10                          // max connections per instance
	config.MinConns = 2                           // keep 2 warm to avoid cold-start latency
	config.MaxConnLifetime = 30 * time.Minute     // recycle connections periodically
	config.MaxConnIdleTime = 5 * time.Minute      // release idle connections quickly (Neon charges for them)
	config.HealthCheckPeriod = 30 * time.Second   // detect stale connections

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	slog.Info("connected to database",
		"maxConns", config.MaxConns,
		"minConns", config.MinConns,
		"maxConnLifetime", config.MaxConnLifetime,
		"maxConnIdleTime", config.MaxConnIdleTime,
	)
	return pool, nil
}
