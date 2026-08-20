// Command encrypt-payloads encrypts webhook payloads that were stored before
// payload encryption was introduced (migration 046).
//
// New payloads are encrypted on write, and old plaintext rows are otherwise
// only removed when they age past the retention window. This backfills them in
// place so the retention window is not the only thing protecting them.
//
// Usage:
//
//	go run ./cmd/encrypt-payloads -dry-run   # report what would change
//	go run ./cmd/encrypt-payloads            # perform the backfill
//
// Prerequisites: TARAN_DB_URL and TARAN_ENCRYPTION_KEY, from the environment or
// backend/.env. The key must be the same one the server runs with; encrypting
// with a different key would leave the rows unreadable to the application.
//
// The run is idempotent and safe to repeat: already-encrypted rows are skipped,
// each row is verified to decrypt back to its original bytes before being
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

func main() {
	dryRun := flag.Bool("dry-run", false, "report what would change without writing")
	batchSize := flag.Int("batch", 100, "rows to read per batch")
	flag.Parse()

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

	repo := database.NewWebhookPayloadRepo(pool, encryptor)

	remaining, err := repo.CountUnencrypted(ctx)
	if err != nil {
		log.Fatalf("count unencrypted payloads: %v", err)
	}
	if remaining == 0 {
		fmt.Println("Nothing to do: no plaintext webhook payloads remain.")
		return
	}

	if *dryRun {
		fmt.Printf("Dry run: %d plaintext payload(s) would be encrypted.\n\n", remaining)
	} else {
		fmt.Printf("Encrypting %d plaintext payload(s)...\n\n", remaining)
	}

	stats, err := repo.BackfillEncryption(ctx, *batchSize, *dryRun, func(s database.BackfillStats) {
		fmt.Printf("  scanned %d, encrypted %d, empty %d, failed %d\n",
			s.Scanned, s.Encrypted, s.SkippedEmpty, s.Failed)
	})
	if err != nil {
		log.Fatalf("backfill failed after %d row(s): %v", stats.Scanned, err)
	}

	verb := "encrypted"
	if *dryRun {
		verb = "would encrypt"
	}
	fmt.Printf("\nDone: %s %d payload(s); %d empty, %d failed, %d scanned.\n",
		verb, stats.Encrypted, stats.SkippedEmpty, stats.Failed, stats.Scanned)

	if stats.Failed > 0 {
		fmt.Fprintf(os.Stderr,
			"\n%d row(s) could not be encrypted and remain in plaintext; see the logged payload IDs.\n",
			stats.Failed)
		os.Exit(1)
	}
}
