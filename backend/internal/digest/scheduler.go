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
)

const retryDelay = 5 * time.Minute

type Scheduler struct {
	generator         *Generator
	emails            database.EmailRepository
	digests           database.DigestRepository
	preferences       database.PreferenceRepository
	sessions          database.SessionRepository
	mailer            mailer.Mailer
	baseURL           string
	unsubscribeSecret string
}

func NewScheduler(generator *Generator, emails database.EmailRepository, digests database.DigestRepository, preferences database.PreferenceRepository, sessions database.SessionRepository, m mailer.Mailer, baseURL, unsubscribeSecret string) *Scheduler {
	return &Scheduler{
		generator:         generator,
		emails:            emails,
		digests:           digests,
		preferences:       preferences,
		sessions:          sessions,
		mailer:            m,
		baseURL:           baseURL,
		unsubscribeSecret: unsubscribeSecret,
	}
}

// RunNow triggers digest generation for all eligible users and
// attempts to send any orphaned unsent digests from previous runs.
// Called by the cron HTTP endpoint (Cloud Scheduler).
func (s *Scheduler) RunNow() {
	s.sendOrphanedDigests()
	s.generateDigests()
}

// sendOrphanedDigests finds digests that were generated but never emailed
// (e.g. due to a crash or transient failure) and attempts to send them.
func (s *Scheduler) sendOrphanedDigests() {
	if s.mailer == nil {
		return
	}

	ctx := context.Background()
	// Look for digests generated more than 10 minutes ago that were never sent.
	olderThan := time.Now().Add(-10 * time.Minute)
	unsent, err := s.digests.ListUnsent(ctx, olderThan, 50)
	if err != nil {
		slog.Error("failed to list unsent digests", "error", err)
		return
	}
	if len(unsent) == 0 {
		return
	}

	slog.Info("found orphaned unsent digests", "count", len(unsent))
	sent := 0
	for i := range unsent {
		d := &unsent[i]

		pref, err := s.preferences.Get(ctx, d.UserID)
		if err != nil {
			slog.Error("failed to get preferences for orphaned digest", "userID", d.UserID, "error", err)
			continue
		}
		if !pref.DigestEmail {
			continue
		}

		email, err := s.sessions.GetUserEmail(ctx, d.UserID)
		if err != nil {
			slog.Error("failed to get email for orphaned digest", "userID", d.UserID, "error", err)
			continue
		}

		// Ensure share token exists for "view in browser" link
		if d.ShareToken == nil {
			tokenBytes := make([]byte, 16)
			if _, err := rand.Read(tokenBytes); err != nil {
				slog.Error("failed to generate share token", "digestID", d.ID, "error", err)
			} else {
				token := hex.EncodeToString(tokenBytes)
				if err := s.digests.SetShareToken(ctx, d.ID, d.UserID, token); err != nil {
					slog.Error("failed to set share token", "digestID", d.ID, "error", err)
				} else {
					d.ShareToken = &token
				}
			}
		}

		var unsubURL string
		if s.baseURL != "" && s.unsubscribeSecret != "" {
			unsubURL = mailer.GenerateUnsubscribeURL(s.baseURL, d.UserID, s.unsubscribeSecret)
		}

		if err := s.mailer.SendDigest(ctx, email, "", d, unsubURL); err != nil {
			slog.Error("failed to send orphaned digest", "digestID", d.ID, "error", err)
			continue
		}

		now := time.Now()
		if err := s.digests.SetSentAt(ctx, d.ID, now); err != nil {
			slog.Error("failed to set sent_at for orphaned digest", "digestID", d.ID, "error", err)
		}
		sent++
	}

	slog.Info("orphaned digest sending complete", "attempted", len(unsent), "sent", sent)
}

// failedUser tracks a user whose digest generation or sending failed so we can retry.
type failedUser struct {
	userID string
	pref   *domain.UserPreference
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
	var failures []failedUser
	for _, userID := range userIDs {
		pref, err := s.preferences.Get(ctx, userID)
		if err != nil {
			slog.Error("failed to get user preferences", "userID", userID, "error", err)
			continue
		}

		if !shouldGenerateForUser(pref, nowUTC) {
			continue
		}

		ok, didGenerate, didSend := s.generateAndSendForUser(ctx, userID, pref, nowUTC)
		if didGenerate {
			generated++
		}
		if didSend {
			sent++
		}
		if !ok {
			failures = append(failures, failedUser{userID: userID, pref: pref})
		}
	}

	slog.Info("digest generation complete", "users_checked", len(userIDs), "generated", generated, "sent", sent, "failures", len(failures))

	if len(failures) > 0 {
		slog.Info("scheduling retry for failed digests", "count", len(failures), "delay", retryDelay)
		time.AfterFunc(retryDelay, func() {
			s.retryFailedDigests(failures, nowUTC)
		})
	}
}

// generateAndSendForUser generates a digest and sends it for a single user.
// Returns (ok, didGenerate, didSend). ok is false if generation or sending failed and should be retried.
func (s *Scheduler) generateAndSendForUser(ctx context.Context, userID string, pref *domain.UserPreference, nowUTC time.Time) (ok, didGenerate, didSend bool) {
	periodStart, periodEnd := computePeriod(pref, nowUTC)

	digest, err := s.generator.GenerateForUser(ctx, userID, pref.DigestFrequency, periodStart, periodEnd)
	if err != nil {
		slog.Error("failed to generate digest", "userID", userID, "error", err)
		return false, false, false
	}
	if digest == nil {
		return true, false, false
	}

	if s.mailer == nil || !pref.DigestEmail {
		return true, true, false
	}

	email, err := s.sessions.GetUserEmail(ctx, userID)
	if err != nil {
		slog.Error("failed to get user email", "userID", userID, "error", err)
		return false, true, false
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
		return false, true, false
	}

	now := time.Now()
	if err := s.digests.SetSentAt(ctx, digest.ID, now); err != nil {
		slog.Error("failed to set digest sent_at", "digestID", digest.ID, "error", err)
	}
	return true, true, true
}

func (s *Scheduler) retryFailedDigests(failures []failedUser, originalNowUTC time.Time) {
	ctx := context.Background()
	slog.Info("retrying failed digests", "count", len(failures))

	retried := 0
	succeeded := 0
	for _, f := range failures {
		ok, _, _ := s.generateAndSendForUser(ctx, f.userID, f.pref, originalNowUTC)
		retried++
		if ok {
			succeeded++
		} else {
			slog.Warn("digest retry also failed", "userID", f.userID)
		}
	}

	slog.Info("digest retry complete", "retried", retried, "succeeded", succeeded)
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

	// Weekly: only generate on the user's chosen day
	if pref.DigestFrequency == "weekly" && userNow.Weekday() != time.Weekday(pref.DigestDay) {
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
