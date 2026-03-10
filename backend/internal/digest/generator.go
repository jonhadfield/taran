package digest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
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

// ErrTokenLimitExceeded is returned when a user has exceeded their monthly token limit.
var ErrTokenLimitExceeded = errors.New("monthly token limit exceeded")

type Generator struct {
	Emails      database.EmailRepository
	Extractions database.ExtractionRepository
	Digests     database.DigestRepository
	Accounts    database.AccountRepository
	Provider    llm.Provider
	SenderPrefs database.SenderPreferenceRepository
	Feedback    database.FeedbackRepository
	Preferences database.PreferenceRepository
	TokenUsage  database.TokenUsageRepository
}

// filteredResult holds the output of the shared filtering pipeline.
type filteredResult struct {
	extractions []domain.Extraction
	emailMap    map[string]*domain.Email
	digestOpts  *llm.DigestOptions
}

// filterExtractions runs the full filtering pipeline: loads extractions for the
// period, removes muted/blocked senders, applies feedback-based filtering,
// keyword exclusions, and sorts by priority. Shared by GenerateForUser and PreviewForUser.
func (g *Generator) filterExtractions(ctx context.Context, userID string, periodStart, periodEnd time.Time) (*filteredResult, error) {
	// Load user's excluded categories (or use defaults)
	var excludedCategories []string
	if g.Preferences != nil {
		pref, err := g.Preferences.Get(ctx, userID)
		if err == nil && len(pref.ExcludedCategories) > 0 {
			excludedCategories = pref.ExcludedCategories
		}
	}

	extractions, err := g.Extractions.ListByUserAndPeriod(ctx, userID, periodStart, periodEnd, excludedCategories...)
	if err != nil {
		return nil, fmt.Errorf("list extractions: %w", err)
	}

	if len(extractions) == 0 {
		return nil, nil
	}

	// Batch-fetch all emails referenced by extractions (single query)
	emailMap := make(map[string]*domain.Email, len(extractions))
	if g.Emails != nil {
		emailIDs := make([]string, 0, len(extractions))
		for _, ext := range extractions {
			emailIDs = append(emailIDs, ext.EmailID)
		}
		fetchedEmails, err := g.Emails.GetByIDsInternal(ctx, emailIDs)
		if err != nil {
			slog.Warn("failed to batch-fetch emails, falling back to empty map", "error", err)
		}
		for i := range fetchedEmails {
			emailMap[fetchedEmails[i].ID] = &fetchedEmails[i]
		}
	}

	// Filter out extractions from muted/blocked senders, apply category overrides, and boost favorites
	if g.SenderPrefs != nil {
		prefs, err := g.SenderPrefs.ListByUser(ctx, userID)
		if err != nil {
			slog.Warn("failed to load sender preferences, proceeding without filter", "error", err)
		} else {
			excluded := make(map[string]bool)
			favorites := make(map[string]bool)
			categoryOverrides := make(map[string]string)
			for _, p := range prefs {
				switch p.Status {
				case "muted", "blocked":
					excluded[p.FromAddress] = true
				case "favorite":
					favorites[p.FromAddress] = true
				}
				if p.Category != "" {
					categoryOverrides[p.FromAddress] = p.Category
				}
			}

			// Build emailID → fromAddress map from pre-fetched emails
			emailAddr := make(map[string]string, len(extractions))
			for _, ext := range extractions {
				if em, ok := emailMap[ext.EmailID]; ok {
					emailAddr[ext.EmailID] = em.FromAddress
				}
			}

			if len(excluded) > 0 {
				var filtered []domain.Extraction
				for _, ext := range extractions {
					addr, ok := emailAddr[ext.EmailID]
					if !ok {
						filtered = append(filtered, ext) // include if we can't determine sender
						continue
					}
					if !excluded[addr] {
						filtered = append(filtered, ext)
					}
				}
				extractions = filtered
				if len(extractions) == 0 {
					return nil, nil
				}
			}

			// Apply sender category overrides: if a sender has a user-assigned category
			// that is in the excluded list, remove those extractions too.
			if len(categoryOverrides) > 0 && len(excludedCategories) > 0 {
				excludedSet := make(map[string]bool, len(excludedCategories))
				for _, c := range excludedCategories {
					excludedSet[c] = true
				}
				var filtered []domain.Extraction
				for _, ext := range extractions {
					addr := emailAddr[ext.EmailID]
					if cat, ok := categoryOverrides[addr]; ok && excludedSet[cat] {
						continue
					}
					filtered = append(filtered, ext)
				}
				extractions = filtered
				if len(extractions) == 0 {
					return nil, nil
				}
			}

			// Sort favorites to front so they get prominence in the digest
			if len(favorites) > 0 {
				sort.SliceStable(extractions, func(i, j int) bool {
					fi := favorites[emailAddr[extractions[i].EmailID]]
					fj := favorites[emailAddr[extractions[j].EmailID]]
					return fi && !fj
				})
			}
		}
	}

	// Apply feedback-based filtering and reordering
	var digestOpts *llm.DigestOptions
	if g.Feedback != nil {
		digestOpts = g.applyFeedback(ctx, userID, &extractions, emailMap)
	}

	// Apply user preferences (digest style + keyword filtering)
	if g.Preferences != nil {
		pref, err := g.Preferences.Get(ctx, userID)
		if err == nil {
			if digestOpts == nil {
				digestOpts = &llm.DigestOptions{}
			}
			if pref.DigestStyle != "" {
				digestOpts.Style = pref.DigestStyle
			}

			// Hard-filter extractions matching exclusion keywords
			if len(pref.ExclusionKeywords) > 0 {
				var filtered []domain.Extraction
				for _, ext := range extractions {
					if !matchesKeywords(ext, pref.ExclusionKeywords) {
						filtered = append(filtered, ext)
					}
				}
				extractions = filtered
				if len(extractions) == 0 {
					return nil, nil
				}
			}

			// Sort interest-matching extractions to front
			if len(pref.InterestKeywords) > 0 {
				sort.SliceStable(extractions, func(i, j int) bool {
					mi := matchesKeywords(extractions[i], pref.InterestKeywords)
					mj := matchesKeywords(extractions[j], pref.InterestKeywords)
					return mi && !mj
				})
			}

			digestOpts.InterestKeywords = pref.InterestKeywords
			digestOpts.ExclusionKeywords = pref.ExclusionKeywords
		}
	}

	return &filteredResult{
		extractions: extractions,
		emailMap:    emailMap,
		digestOpts:  digestOpts,
	}, nil
}

// PreviewForUser returns which emails would be included in a digest
// after all filtering, without calling the LLM or persisting anything.
func (g *Generator) PreviewForUser(ctx context.Context, userID string, periodType string, periodStart, periodEnd time.Time) (*domain.DigestPreview, error) {
	result, err := g.filterExtractions(ctx, userID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	preview := &domain.DigestPreview{
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		PeriodType:  periodType,
		EmailCount:  len(result.extractions),
	}

	for _, ext := range result.extractions {
		item := domain.DigestPreviewItem{
			EmailID: ext.EmailID,
			Summary: ext.Summary,
		}
		if em, ok := result.emailMap[ext.EmailID]; ok {
			item.Subject = em.Subject
			item.FromName = em.FromName
			item.FromAddress = em.FromAddress
			item.ReceivedAt = em.ReceivedAt
		}
		preview.Items = append(preview.Items, item)
	}

	return preview, nil
}

func (g *Generator) GenerateForUser(ctx context.Context, userID string, periodType string, periodStart, periodEnd time.Time) (*domain.Digest, error) {
	// Skip if a digest already exists for this exact period (dedup against concurrent triggers)
	if g.Digests != nil {
		exists, err := g.Digests.ExistsForPeriod(ctx, userID, periodStart, periodEnd)
		if err != nil {
			return nil, fmt.Errorf("check existing digest: %w", err)
		}
		if exists {
			slog.Info("digest already exists for period, skipping", "userID", userID, "periodStart", periodStart, "periodEnd", periodEnd)
			return nil, nil
		}
	}

	// Check monthly token limit before making LLM call
	if g.TokenUsage != nil && g.Preferences != nil {
		pref, err := g.Preferences.Get(ctx, userID)
		if err == nil && pref.MonthlyTokenLimit > 0 {
			used, err := g.TokenUsage.GetMonthlyTotal(ctx, userID)
			if err != nil {
				slog.Warn("failed to check token limit, proceeding", "error", err)
			} else if used >= pref.MonthlyTokenLimit {
				slog.Warn("user exceeded monthly token limit",
					"userID", userID, "used", used, "limit", pref.MonthlyTokenLimit)
				return nil, ErrTokenLimitExceeded
			}
		}
	}

	result, err := g.filterExtractions(ctx, userID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	extractions := result.extractions
	emailMap := result.emailMap

	summary, usage, err := g.Provider.GenerateDigest(ctx, extractions, periodType, result.digestOpts)
	if err != nil {
		return nil, fmt.Errorf("generate digest summary: %w", err)
	}

	// Record digest token usage
	if g.TokenUsage != nil && usage != nil {
		tu := &domain.TokenUsage{
			ID:           uuid.New().String(),
			UserID:       userID,
			Operation:    "digest",
			Provider:     g.Provider.Name(),
			Model:        g.Provider.Model(),
			InputTokens:  usage.InputTokens,
			OutputTokens: usage.OutputTokens,
			TotalTokens:  usage.TotalTokens,
			CreatedAt:    time.Now(),
		}
		if err := g.TokenUsage.Create(ctx, tu); err != nil {
			slog.Warn("failed to record digest token usage", "error", err)
		}
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
		TokensUsed:  usage.TotalTokens,
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

	// Build email summaries from pre-fetched emailMap
	for _, ext := range extractions {
		if em, ok := emailMap[ext.EmailID]; ok {
			digest.EmailSummaries = append(digest.EmailSummaries, domain.DigestEmailSummary{
				EmailID:    ext.EmailID,
				Subject:    em.Subject,
				SenderName: em.FromName,
				Summary:    ext.Summary,
			})
		}
	}

	if err := g.Digests.Create(ctx, digest); err != nil {
		if errors.Is(err, database.ErrDigestDuplicate) {
			slog.Info("digest dedup: concurrent insert detected, skipping",
				"userID", userID, "periodStart", periodStart, "periodEnd", periodEnd)
			return nil, nil
		}
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
func (g *Generator) applyFeedback(ctx context.Context, userID string, extractions *[]domain.Extraction, emailMap map[string]*domain.Email) *llm.DigestOptions {
	// Build emailID → fromAddress map from pre-fetched emails
	fromAddr := make(map[string]string, len(*extractions))
	for _, ext := range *extractions {
		if em, ok := emailMap[ext.EmailID]; ok {
			fromAddr[ext.EmailID] = em.FromAddress
		}
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

// matchesKeywords returns true if any keyword matches one of the extraction's
// topics (case-insensitive exact match) or appears in its summary (case-insensitive substring).
func matchesKeywords(ext domain.Extraction, keywords []string) bool {
	for _, kw := range keywords {
		kwLower := strings.ToLower(kw)
		for _, topic := range ext.Topics {
			if strings.EqualFold(topic, kw) {
				return true
			}
		}
		if strings.Contains(strings.ToLower(ext.Summary), kwLower) {
			return true
		}
	}
	return false
}
