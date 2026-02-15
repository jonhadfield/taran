package digest

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/llm"
)

const (
	minSenderFeedbackCount      = 3
	senderExcludeThreshold      = 0.75
	minTopicFeedbackCount       = 3
	topicPreferredThreshold     = 0.65
	topicLessPreferredThreshold = 0.35
)

type Generator struct {
	Emails      database.EmailRepository
	Extractions database.ExtractionRepository
	Digests     database.DigestRepository
	Accounts    database.AccountRepository
	Provider    llm.Provider
	SenderPrefs database.SenderPreferenceRepository
	Feedback    database.FeedbackRepository
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

	// Apply feedback-based filtering and reordering
	var digestOpts *llm.DigestOptions
	if g.Feedback != nil {
		digestOpts = g.applyFeedback(ctx, userID, &extractions)
	}

	summary, usage, err := g.Provider.GenerateDigest(ctx, extractions, periodType, digestOpts)
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

// applyFeedback filters out poorly-rated senders, reorders extractions by
// sender quality, and builds topic preferences for the LLM prompt.
// It modifies the extractions slice in place via the pointer.
func (g *Generator) applyFeedback(ctx context.Context, userID string, extractions *[]domain.Extraction) *llm.DigestOptions {
	// Build emailID → fromAddress map in a single pass
	fromAddr := make(map[string]string, len(*extractions))
	for _, ext := range *extractions {
		email, err := g.Emails.GetByIDInternal(ctx, ext.EmailID)
		if err != nil || email == nil {
			continue
		}
		fromAddr[ext.EmailID] = email.FromAddress
	}

	// Sender filtering and reordering
	senderStats, err := g.Feedback.GetSenderStats(ctx, userID)
	if err != nil {
		slog.Warn("failed to load sender feedback stats, skipping feedback filter", "error", err)
		return nil
	}

	excludedSenders := make(map[string]bool)
	senderRatio := make(map[string]float64)
	for _, s := range senderStats {
		total := s.UsefulCount + s.NotUsefulCount
		ratio := float64(s.NotUsefulCount) / float64(total)
		if total >= minSenderFeedbackCount && ratio > senderExcludeThreshold {
			excludedSenders[s.FromAddress] = true
		} else {
			senderRatio[s.FromAddress] = float64(s.UsefulCount) / float64(total)
		}
	}

	if len(excludedSenders) > 0 {
		filtered := make([]domain.Extraction, 0, len(*extractions))
		for _, ext := range *extractions {
			if addr, ok := fromAddr[ext.EmailID]; ok && excludedSenders[addr] {
				continue
			}
			filtered = append(filtered, ext)
		}
		*extractions = filtered
		if len(*extractions) == 0 {
			return nil
		}
	}

	// Stable sort: positively-rated senders first
	sort.SliceStable(*extractions, func(i, j int) bool {
		addrI := fromAddr[(*extractions)[i].EmailID]
		addrJ := fromAddr[(*extractions)[j].EmailID]
		return senderRatio[addrI] > senderRatio[addrJ]
	})

	// Topic enrichment
	topicStats, err := g.Feedback.GetTopicStats(ctx, userID)
	if err != nil {
		slog.Warn("failed to load topic feedback stats, skipping topic preferences", "error", err)
		return nil
	}

	var opts llm.DigestOptions
	for _, t := range topicStats {
		total := t.UsefulCount + t.NotUsefulCount
		if total < minTopicFeedbackCount {
			continue
		}
		ratio := float64(t.UsefulCount) / float64(total)
		if ratio >= topicPreferredThreshold {
			opts.PreferredTopics = append(opts.PreferredTopics, t.Topic)
		} else if ratio <= topicLessPreferredThreshold {
			opts.LessPreferredTopics = append(opts.LessPreferredTopics, t.Topic)
		}
	}

	if len(opts.PreferredTopics) == 0 && len(opts.LessPreferredTopics) == 0 {
		return nil
	}
	return &opts
}
