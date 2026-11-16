package database

import (
	"context"
	"log"
	"os"
	"testing"
)

// TestMain prepares a shared Postgres schema when TARAN_TEST_DB_URL is set so
// integration tests (encryption backfill, preference batch) can run in CI.
func TestMain(m *testing.M) {
	dsn := os.Getenv("TARAN_TEST_DB_URL")
	if dsn == "" {
		os.Exit(m.Run())
	}

	ctx := context.Background()
	pool, err := NewPool(ctx, dsn)
	if err != nil {
		log.Fatalf("test db connect: %v", err)
	}
	defer pool.Close()

	// Auth tables are owned by Prisma; create a minimal "user" so Go migrations
	// that FK to it can apply on a fresh database.
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS "user" (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL DEFAULT '',
			email TEXT NOT NULL DEFAULT '',
			"emailVerified" BOOLEAN NOT NULL DEFAULT FALSE,
			image TEXT,
			"createdAt" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			"updatedAt" TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`); err != nil {
		log.Fatalf("create stub user table: %v", err)
	}

	if err := RunMigrations(ctx, pool); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	os.Exit(m.Run())
}
