// Command backfill-encryption encrypts data that was stored before encryption
// at rest was enabled.
//
// Email bodies and webhook payloads are encrypted on write whenever
// TARAN_ENCRYPTION_KEY is configured, but rows written before that point keep
// their plaintext. Webhook payloads are eventually removed by the retention
// sweep; email bodies are kept indefinitely. This backfills both in place.
//
// Usage:
//
//	go run ./cmd/backfill-encryption -dry-run           # report what would change
//	go run ./cmd/backfill-encryption                    # backfill everything
//	go run ./cmd/backfill-encryption -target emails     # emails only
//	go run ./cmd/backfill-encryption -target payloads   # webhook payloads only
//
// Prerequisites: TARAN_DB_URL and TARAN_ENCRYPTION_KEY, from the environment or
// backend/.env. The key must be the same one the server runs with; encrypting
// with a different key would leave the rows unreadable to the application.
//
// The run is idempotent and safe to repeat: already-encrypted rows are skipped,
// every value is verified to decrypt back to its original before being
// overwritten, and any row that fails is left in plaintext and reported.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hadfielj/taran/backend/internal/crypto"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/joho/godotenv"
)

// backfiller is the shared shape of the per-table backfills.
type backfiller struct {
	name    string
	count   func(context.Context) (int, error)
	migrate func(context.Context, int, bool, func(database.BackfillStats)) (database.BackfillStats, error)
}

func main() {
	dryRun := flag.Bool("dry-run", false, "report what would change without writing")
	batchSize := flag.Int("batch", 100, "rows to read per batch")
	target := flag.String("target", "all", "what to backfill: all, emails, or payloads")
	flag.Parse()

	if *target != "all" && *target != "emails" && *target != "payloads" {
		log.Fatalf("invalid -target %q: want all, emails, or payloads", *target)
	}

	_ = godotenv.Load()

	dbURL := os.Getenv("TARAN_DB_URL")
	if dbURL == "" {
		log.Fatal("TARAN_DB_URL is required")
	}
	keyHex := os.Getenv("TARAN_ENCRYPTION_KEY")
	if keyHex == "" {
		log.Fatal("TARAN_ENCRYPTION_KEY is required; it must match the key the server runs with")
	}

	encryptor, err := crypto.NewEncryptor(keyHex)
	if err != nil {
		log.Fatalf("invalid encryption key: %v", err)
	}

	// Ctrl-C stops between batches; rows already committed stay encrypted and a
	// later run resumes from where this one stopped.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	emails := database.NewEmailRepo(pool, encryptor)
	payloads := database.NewWebhookPayloadRepo(pool, encryptor)

	var jobs []backfiller
	if *target == "all" || *target == "emails" {
		jobs = append(jobs, backfiller{"email bodies", emails.CountUnencrypted, emails.BackfillEncryption})
	}
	if *target == "all" || *target == "payloads" {
		jobs = append(jobs, backfiller{"webhook payloads", payloads.CountUnencrypted, payloads.BackfillEncryption})
	}

	totalFailed := 0
	for _, job := range jobs {
		remaining, err := job.count(ctx)
		if err != nil {
			log.Fatalf("count %s: %v", job.name, err)
		}
		if remaining == 0 {
			fmt.Printf("%s: nothing to do.\n\n", job.name)
			continue
		}

		if *dryRun {
			fmt.Printf("%s: %d row(s) would be encrypted.\n", job.name, remaining)
		} else {
			fmt.Printf("%s: encrypting %d row(s)...\n", job.name, remaining)
		}

		stats, err := job.migrate(ctx, *batchSize, *dryRun, func(s database.BackfillStats) {
			fmt.Printf("  scanned %d, encrypted %d, empty %d, failed %d\n",
				s.Scanned, s.Encrypted, s.SkippedEmpty, s.Failed)
		})
		if err != nil {
			log.Fatalf("%s: backfill failed after %d row(s): %v", job.name, stats.Scanned, err)
		}

		verb := "encrypted"
		if *dryRun {
			verb = "would encrypt"
		}
		fmt.Printf("  done: %s %d, empty %d, failed %d, scanned %d\n\n",
			verb, stats.Encrypted, stats.SkippedEmpty, stats.Failed, stats.Scanned)
		totalFailed += stats.Failed
	}

	if totalFailed > 0 {
		fmt.Fprintf(os.Stderr,
			"%d row(s) could not be encrypted and remain in plaintext; see the logged row IDs.\n",
			totalFailed)
		os.Exit(1)
	}
}
