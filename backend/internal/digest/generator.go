package digest

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/llm"
)

type Generator struct {
	Emails      database.EmailRepository
	Extractions database.ExtractionRepository
	Digests     database.DigestRepository
	Accounts    database.AccountRepository
	Provider    llm.Provider
	SenderPrefs database.SenderPreferenceRepository
}

func (g *Generator) GenerateForUser(ctx context.Context, userID string, periodType string, periodStart, periodEnd time.Time) (*domain.Digest, error) {
	extractions, err := g.Extractions.ListByUserAndPeriod(ctx, userID, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("list extractions: %w", err)
	}

	if len(extractions) == 0 {
		return nil, nil
	}

	// Filter out extractions from muted/blocked senders
	if g.SenderPrefs != nil {
		prefs, err := g.SenderPrefs.ListByUser(ctx, userID)
		if err != nil {
			slog.Warn("failed to load sender preferences, proceeding without filter", "error", err)
		} else {
			excluded := make(map[string]bool)
			for _, p := range prefs {
				if p.Status == "muted" || p.Status == "blocked" {
					excluded[p.FromAddress] = true
				}
			}
			if len(excluded) > 0 {
				var filtered []domain.Extraction
				for _, ext := range extractions {
					email, err := g.Emails.GetByIDInternal(ctx, ext.EmailID)
					if err != nil || email == nil {
						filtered = append(filtered, ext) // include if we can't determine sender
						continue
					}
					if !excluded[email.FromAddress] {
						filtered = append(filtered, ext)
					}
				}
				extractions = filtered
				if len(extractions) == 0 {
					return nil, nil
				}
			}
		}
	}

	summary, usage, err := g.Provider.GenerateDigest(ctx, extractions, periodType)
	if err != nil {
		return nil, fmt.Errorf("generate digest summary: %w", err)
	}

	now := time.Now()
	digest := &domain.Digest{
		ID:          uuid.New().String(),
		UserID:      userID,
		Title:       summary.Title,
		Summary:     summary.Summary,
		Highlights:  summary.Highlights,
		TopTopics:   summary.TopTopics,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		PeriodType:  periodType,
		EmailCount:  len(extractions),
		Provider:    g.Provider.Name(),
		Model:       g.Provider.Model(),
		GeneratedAt: now,
		CreatedAt:   now,
	}

	for i, ext := range extractions {
		digest.Items = append(digest.Items, domain.DigestItem{
			ID:           uuid.New().String(),
			DigestID:     digest.ID,
			EmailID:      ext.EmailID,
			ExtractionID: ext.ID,
			SortOrder:    i,
		})
	}

	// Build email summaries for email rendering
	for _, ext := range extractions {
		email, err := g.Emails.GetByIDInternal(ctx, ext.EmailID)
		if err != nil || email == nil {
			continue
		}
		digest.EmailSummaries = append(digest.EmailSummaries, domain.DigestEmailSummary{
			EmailID:    ext.EmailID,
			Subject:    email.Subject,
			SenderName: email.FromName,
			Summary:    ext.Summary,
		})
	}

	if err := g.Digests.Create(ctx, digest); err != nil {
		return nil, fmt.Errorf("store digest: %w", err)
	}

	slog.Info("digest generated",
		"userID", userID,
		"period", periodType,
		"emails", len(extractions),
		"tokens", usage.TotalTokens,
	)

	return digest, nil
}
