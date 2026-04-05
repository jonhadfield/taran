// Seed populates the local development database with realistic test data.
// Usage: go run ./cmd/seed/
//
// Prerequisites:
//   - PostgreSQL running (make db)
//   - At least one user must exist (sign in via the frontend first)
//   - Backend .env file with TARAN_DB_URL set
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var senders = []struct {
	name, address, category string
}{
	{"Morning Brew", "morning@morningbrew.com", "newsletter"},
	{"TLDR", "dan@tldrnewsletter.com", "newsletter"},
	{"Hacker News Digest", "noreply@hndigest.com", "newsletter"},
	{"The Hustle", "sam@thehustle.co", "newsletter"},
	{"Stratechery", "ben@stratechery.com", "newsletter"},
	{"Platformer", "casey@platformer.news", "newsletter"},
	{"GitHub", "noreply@github.com", "notification"},
	{"Stripe", "receipts@stripe.com", "transactional"},
	{"AWS", "no-reply@sns.amazonaws.com", "notification"},
	{"Alice Johnson", "alice.johnson@gmail.com", "personal"},
}

var subjects = []string{
	"The Future of AI in Enterprise Software",
	"Weekly Roundup: Top Stories in Tech",
	"Your Monday Briefing: Markets, AI, and More",
	"How to Build a Scalable Data Pipeline",
	"The Rise of Vertical SaaS",
	"Breaking: Major Acquisition Announced",
	"Deep Dive: Kubernetes Best Practices",
	"Newsletter: What We Read This Week",
	"5 Things You Missed in Tech This Week",
	"The State of Open Source in 2026",
	"Product Update: New Features Released",
	"Engineering Blog: Lessons from Production",
	"Weekly Digest: Security & Infrastructure",
	"The Developer Experience Gap",
	"How We Scaled to 1M Users",
}

var summaries = []string{
	"A comprehensive overview of how AI is transforming enterprise workflows, from automated code review to intelligent customer support systems.",
	"This week's top stories include a major funding round, a new open-source project gaining traction, and regulatory changes affecting tech companies.",
	"Markets opened higher on positive earnings reports. AI chip demand continues to outpace supply. Several startups announced pivots to AI-first strategies.",
	"A practical guide to building data pipelines that handle millions of events per day using modern streaming architectures.",
	"Analysis of how vertical SaaS companies are capturing market share by deeply understanding specific industry needs.",
	"Coverage of the latest major acquisition in the tech sector and its implications for competition and innovation.",
	"Best practices for running Kubernetes in production, including resource management, monitoring, and disaster recovery.",
	"A curated collection of the most interesting articles, papers, and discussions from the past week.",
	"Highlights from the tech industry including product launches, funding rounds, and key personnel moves.",
	"A detailed look at the current state of open source, including sustainability challenges and new governance models.",
}

var topics = [][]string{
	{"AI", "Enterprise", "Machine Learning"},
	{"Tech News", "Startups", "Funding"},
	{"Markets", "AI", "Business"},
	{"Data Engineering", "Architecture", "Scalability"},
	{"SaaS", "Business Model", "Industry"},
	{"M&A", "Competition", "Strategy"},
	{"Kubernetes", "DevOps", "Infrastructure"},
	{"Reading", "Curation", "Knowledge"},
	{"Tech News", "Product", "Innovation"},
	{"Open Source", "Community", "Sustainability"},
}

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("TARAN_DB_URL")
	if dbURL == "" {
		log.Fatal("TARAN_DB_URL environment variable required")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	// Find the first user
	var userID string
	err = pool.QueryRow(ctx, `SELECT id FROM "user" LIMIT 1`).Scan(&userID)
	if err != nil {
		log.Fatal("no users found — sign in via the frontend first, then run this script")
	}
	fmt.Printf("Seeding data for user %s\n", userID)

	// Find user's email account
	var accountID, emailAddress string
	err = pool.QueryRow(ctx, `SELECT id, email_address FROM email_account WHERE user_id = $1 LIMIT 1`, userID).Scan(&accountID, &emailAddress)
	if err != nil {
		log.Fatal("no email account found — complete onboarding first")
	}
	fmt.Printf("Using account %s (%s)\n", accountID, emailAddress)

	now := time.Now()
	emailIDs := make([]string, 0, 30)

	// Create 30 emails spread over the last 14 days
	for i := 0; i < 30; i++ {
		sender := senders[i%len(senders)]
		subject := subjects[i%len(subjects)]
		summary := summaries[i%len(summaries)]
		topicList := topics[i%len(topics)]

		daysAgo := rand.IntN(14)
		hoursAgo := rand.IntN(24)
		receivedAt := now.Add(-time.Duration(daysAgo)*24*time.Hour - time.Duration(hoursAgo)*time.Hour)

		emailID := uuid.New().String()
		emailIDs = append(emailIDs, emailID)

		_, err := pool.Exec(ctx, `
			INSERT INTO email (id, user_id, account_id, message_id, from_address, from_name, to_address, subject, text_body, received_at, date_header, status, is_read, is_starred, is_archived)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10, 'processed', $11, $12, false)
			ON CONFLICT DO NOTHING`,
			emailID, userID, accountID,
			fmt.Sprintf("<%s@seed>", emailID),
			sender.address, sender.name, emailAddress, subject,
			fmt.Sprintf("This is the text body for: %s\n\n%s", subject, summary),
			receivedAt,
			i >= 10, // first 10 are unread
			i%7 == 0, // every 7th is starred
		)
		if err != nil {
			log.Printf("insert email %d: %v", i, err)
			continue
		}

		// Create extraction
		extractionID := uuid.New().String()
		_, err = pool.Exec(ctx, `
			INSERT INTO extraction (id, email_id, summary, key_points, topics, links, action_items, sentiment, source_category, provider, model, tokens_used, processed_at)
			VALUES ($1, $2, $3, $4, $5, '[]'::jsonb, $6, $7, $8, 'seed', 'seed', 0, $9)
			ON CONFLICT DO NOTHING`,
			extractionID, emailID, summary,
			topicList[:min(3, len(topicList))], // key_points as string array
			topicList,
			[]string{}, // action_items
			"informational",
			sender.category,
			receivedAt,
		)
		if err != nil {
			log.Printf("insert extraction %d: %v", i, err)
		}
	}

	// Create 2 digests
	for i := 0; i < 2; i++ {
		digestID := uuid.New().String()
		periodEnd := now.Add(-time.Duration(i*7) * 24 * time.Hour)
		periodStart := periodEnd.Add(-7 * 24 * time.Hour)

		_, err := pool.Exec(ctx, `
			INSERT INTO digest (id, user_id, title, summary, highlights, top_topics, period_start, period_end, period_type, email_count)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'weekly', $9)
			ON CONFLICT DO NOTHING`,
			digestID, userID,
			fmt.Sprintf("Weekly Digest — %s", periodStart.Format("Jan 2")),
			"A busy week with newsletters covering AI developments, infrastructure best practices, and market trends.",
			[]string{"AI chip demand continues to surge", "New open-source database project gains traction", "Enterprise SaaS consolidation accelerating"},
			[]string{"AI", "Tech News", "Infrastructure", "SaaS"},
			periodStart, periodEnd,
			min(15, len(emailIDs)),
		)
		if err != nil {
			log.Printf("insert digest %d: %v", i, err)
			continue
		}

		// Link some emails to the digest
		itemCount := min(15, len(emailIDs))
		start := i * 15
		for j := 0; j < itemCount && start+j < len(emailIDs); j++ {
			itemID := uuid.New().String()
			pool.Exec(ctx, `
				INSERT INTO digest_item (id, digest_id, email_id, sort_order)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT DO NOTHING`,
				itemID, digestID, emailIDs[start+j], j,
			)
		}
	}

	// Create a few labels
	labelNames := []struct{ name, color string }{
		{"Important", "red"},
		{"Work", "blue"},
		{"Reading List", "green"},
	}
	for _, l := range labelNames {
		labelID := uuid.New().String()
		pool.Exec(ctx, `
			INSERT INTO label (id, user_id, name, color)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT DO NOTHING`,
			labelID, userID, l.name, l.color,
		)
	}

	fmt.Printf("Seeded: 30 emails, 2 digests, 3 labels\n")
	fmt.Println("Done! Refresh the dashboard to see the data.")
}
