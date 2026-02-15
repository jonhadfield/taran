package digest

import (
	"context"
	"log/slog"
	"time"

	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/mailer"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron        *cron.Cron
	generator   *Generator
	emails      database.EmailRepository
	digests     database.DigestRepository
	preferences database.PreferenceRepository
	sessions    database.SessionRepository
	mailer      mailer.Mailer
}

func NewScheduler(cronExpr, timezone string, generator *Generator, emails database.EmailRepository, digests database.DigestRepository, preferences database.PreferenceRepository, sessions database.SessionRepository, m mailer.Mailer) (*Scheduler, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	c := cron.New(cron.WithLocation(loc))

	s := &Scheduler{
		cron:        c,
		generator:   generator,
		emails:      emails,
		digests:     digests,
		preferences: preferences,
		sessions:    sessions,
		mailer:      m,
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
	sent := 0
	for _, userID := range userIDs {
		digest, err := s.generator.GenerateForUser(ctx, userID, "daily", periodStart, periodEnd)
		if err != nil {
			slog.Error("failed to generate digest", "userID", userID, "error", err)
			continue
		}
		if digest == nil {
			continue
		}
		generated++

		if s.mailer == nil {
			continue
		}

		pref, err := s.preferences.Get(ctx, userID)
		if err != nil {
			slog.Error("failed to get user preferences", "userID", userID, "error", err)
			continue
		}
		if !pref.DigestEmail {
			continue
		}

		email, err := s.sessions.GetUserEmail(ctx, userID)
		if err != nil {
			slog.Error("failed to get user email", "userID", userID, "error", err)
			continue
		}

		if err := s.mailer.SendDigest(ctx, email, "", digest); err != nil {
			slog.Error("failed to send digest email", "userID", userID, "error", err)
			continue
		}

		now := time.Now()
		if err := s.digests.SetSentAt(ctx, digest.ID, now); err != nil {
			slog.Error("failed to set digest sent_at", "digestID", digest.ID, "error", err)
		}
		sent++
	}

	slog.Info("daily digest generation complete", "users", len(userIDs), "generated", generated, "sent", sent)
}
