package digest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/mailer"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron              *cron.Cron
	generator         *Generator
	emails            database.EmailRepository
	digests           database.DigestRepository
	preferences       database.PreferenceRepository
	sessions          database.SessionRepository
	mailer            mailer.Mailer
	baseURL           string
	unsubscribeSecret string
}

func NewScheduler(generator *Generator, emails database.EmailRepository, digests database.DigestRepository, preferences database.PreferenceRepository, sessions database.SessionRepository, m mailer.Mailer, baseURL, unsubscribeSecret string) (*Scheduler, error) {
	c := cron.New(cron.WithLocation(time.UTC))

	s := &Scheduler{
		cron:              c,
		generator:         generator,
		emails:            emails,
		digests:           digests,
		preferences:       preferences,
		sessions:          sessions,
		mailer:            m,
		baseURL:           baseURL,
		unsubscribeSecret: unsubscribeSecret,
	}

	// Run every hour on the hour
	_, err := c.AddFunc("0 * * * *", func() {
		s.generateDigests()
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
		slog.Info("digest scheduler started (hourly)", "next_run", entries[0].Next)
	}
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	slog.Info("digest scheduler stopped")
}

func (s *Scheduler) generateDigests() {
	ctx := context.Background()
	nowUTC := time.Now().UTC()

	slog.Info("digest scheduler tick", "utc_hour", nowUTC.Hour())

	// Get all users who had email activity in the last 7 days (covers both daily and weekly)
	lookback := nowUTC.Add(-7 * 24 * time.Hour)
	userIDs, err := s.emails.ListActiveUserIDs(ctx, lookback, nowUTC)
	if err != nil {
		slog.Error("failed to get active users", "error", err)
		return
	}

	generated := 0
	sent := 0
	for _, userID := range userIDs {
		pref, err := s.preferences.Get(ctx, userID)
		if err != nil {
			slog.Error("failed to get user preferences", "userID", userID, "error", err)
			continue
		}

		if !shouldGenerateForUser(pref, nowUTC) {
			continue
		}

		periodStart, periodEnd := computePeriod(pref, nowUTC)

		digest, err := s.generator.GenerateForUser(ctx, userID, pref.DigestFrequency, periodStart, periodEnd)
		if err != nil {
			slog.Error("failed to generate digest", "userID", userID, "error", err)
			continue
		}
		if digest == nil {
			continue
		}
		generated++

		if s.mailer == nil || !pref.DigestEmail {
			continue
		}

		email, err := s.sessions.GetUserEmail(ctx, userID)
		if err != nil {
			slog.Error("failed to get user email", "userID", userID, "error", err)
			continue
		}

		// Auto-generate share token for "view in browser" link
		if digest.ShareToken == nil {
			tokenBytes := make([]byte, 16)
			if _, err := rand.Read(tokenBytes); err != nil {
				slog.Error("failed to generate share token", "digestID", digest.ID, "error", err)
			} else {
				token := hex.EncodeToString(tokenBytes)
				if err := s.digests.SetShareToken(ctx, digest.ID, userID, token); err != nil {
					slog.Error("failed to set share token", "digestID", digest.ID, "error", err)
				} else {
					digest.ShareToken = &token
				}
			}
		}

		var unsubURL string
		if s.baseURL != "" && s.unsubscribeSecret != "" {
			unsubURL = mailer.GenerateUnsubscribeURL(s.baseURL, userID, s.unsubscribeSecret)
		}

		if err := s.mailer.SendDigest(ctx, email, "", digest, unsubURL); err != nil {
			slog.Error("failed to send digest email", "userID", userID, "error", err)
			continue
		}

		now := time.Now()
		if err := s.digests.SetSentAt(ctx, digest.ID, now); err != nil {
			slog.Error("failed to set digest sent_at", "digestID", digest.ID, "error", err)
		}
		sent++
	}

	slog.Info("digest generation complete", "users_checked", len(userIDs), "generated", generated, "sent", sent)
}

// shouldGenerateForUser checks if the current UTC hour matches the user's preferred delivery hour
// in their timezone, and respects the frequency setting.
func shouldGenerateForUser(pref *domain.UserPreference, nowUTC time.Time) bool {
	loc, err := time.LoadLocation(pref.DigestTimezone)
	if err != nil {
		loc = time.UTC
	}

	userNow := nowUTC.In(loc)
	if userNow.Hour() != pref.DigestHour {
		return false
	}

	// Weekly: only generate on Monday
	if pref.DigestFrequency == "weekly" && userNow.Weekday() != time.Monday {
		return false
	}

	return true
}

// computePeriod returns the digest period based on the user's frequency.
func computePeriod(pref *domain.UserPreference, nowUTC time.Time) (time.Time, time.Time) {
	loc, err := time.LoadLocation(pref.DigestTimezone)
	if err != nil {
		loc = time.UTC
	}

	userNow := nowUTC.In(loc)
	periodEnd := time.Date(userNow.Year(), userNow.Month(), userNow.Day(), userNow.Hour(), 0, 0, 0, loc).UTC()

	switch pref.DigestFrequency {
	case "weekly":
		return periodEnd.Add(-7 * 24 * time.Hour), periodEnd
	default:
		return periodEnd.Add(-24 * time.Hour), periodEnd
	}
}
