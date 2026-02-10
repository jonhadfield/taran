package digest

import (
	"context"
	"log/slog"
	"time"

	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron      *cron.Cron
	generator *Generator
	emails    database.EmailRepository
}

func NewScheduler(cronExpr, timezone string, generator *Generator, emails database.EmailRepository) (*Scheduler, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	c := cron.New(cron.WithLocation(loc))

	s := &Scheduler{
		cron:      c,
		generator: generator,
		emails:    emails,
	}

	_, err = c.AddFunc(cronExpr, func() {
		s.generateDailyDigests()
	})
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Scheduler) Start() {
	s.cron.Start()
	entries := s.cron.Entries()
	if len(entries) > 0 {
		slog.Info("digest scheduler started", "next_run", entries[0].Next)
	}
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	slog.Info("digest scheduler stopped")
}

func (s *Scheduler) generateDailyDigests() {
	ctx := context.Background()

	now := time.Now()
	periodEnd := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	periodStart := periodEnd.Add(-24 * time.Hour)

	slog.Info("generating daily digests", "period_start", periodStart, "period_end", periodEnd)

	userIDs, err := s.emails.ListActiveUserIDs(ctx, periodStart, periodEnd)
	if err != nil {
		slog.Error("failed to get active users", "error", err)
		return
	}

	generated := 0
	for _, userID := range userIDs {
		digest, err := s.generator.GenerateForUser(ctx, userID, "daily", periodStart, periodEnd)
		if err != nil {
			slog.Error("failed to generate digest", "userID", userID, "error", err)
			continue
		}
		if digest != nil {
			generated++
		}
	}

	slog.Info("daily digest generation complete", "users", len(userIDs), "generated", generated)
}
