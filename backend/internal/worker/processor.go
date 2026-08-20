package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
	emailpkg "github.com/hadfielj/taran/backend/internal/email"
	"github.com/hadfielj/taran/backend/internal/llm"
	"github.com/hadfielj/taran/backend/internal/mailer"
	"github.com/hadfielj/taran/backend/internal/sse"
)

const (
	extractMaxRetries    = 3
	extractBaseDelay     = 2 * time.Second
	sweepInterval        = 5 * time.Minute
	sweepBatchSize       = 100
	startupRequeueMax    = 500
	retryMaxAttempts     = 5 // max auto-retries for failed emails (backoff: 5m, 20m, 80m, 320m, 1280m)
	retryBatchSize       = 20
	triagePreviewMaxLen  = 1500 // max characters of content sent to triage LLM call
)

type Processor struct {
	queue       chan string
	done        chan struct{}
	emails      database.EmailRepository
	extractions database.ExtractionRepository
	resolver    *llm.ProviderResolver
	senderPrefs database.SenderPreferenceRepository
	tokenUsage  database.TokenUsageRepository
	preferences database.PreferenceRepository
	wg          sync.WaitGroup
	concurrency int

	// Optional deps (set after construction)
	Mailer           mailer.Mailer
	Sessions         database.SessionRepository
	AutoArchiveRules database.AutoArchiveRuleRepository
	SSEBroker        *sse.Broker
	Digests          database.DigestRepository

	// WebhookPayloads and PayloadRetention drive deletion of stored raw
	// messages. Zero retention disables the sweep.
	WebhookPayloads  database.WebhookPayloadRepository
	PayloadRetention time.Duration
}

// ProcessorConfig groups the dependencies for NewProcessor.
type ProcessorConfig struct {
	BufferSize  int
	Concurrency int
	Emails      database.EmailRepository
	Extractions database.ExtractionRepository
	Resolver    *llm.ProviderResolver
	SenderPrefs database.SenderPreferenceRepository
	TokenUsage  database.TokenUsageRepository
	Preferences database.PreferenceRepository
}

func NewProcessor(cfg ProcessorConfig) *Processor {
	return &Processor{
		queue:       make(chan string, cfg.BufferSize),
		done:        make(chan struct{}),
		emails:      cfg.Emails,
		extractions: cfg.Extractions,
		resolver:    cfg.Resolver,
		senderPrefs: cfg.SenderPrefs,
		tokenUsage:  cfg.TokenUsage,
		preferences: cfg.Preferences,
		concurrency: cfg.Concurrency,
	}
}

func (p *Processor) Start(ctx context.Context) {
	pending, err := p.emails.ListPending(ctx, startupRequeueMax)
	if err != nil {
		slog.Error("failed to list pending emails", "error", err)
	} else {
		for _, e := range pending {
			p.Enqueue(e.ID)
		}
		if len(pending) > 0 {
			slog.Info("re-queued pending emails", "count", len(pending))
		}
	}

	for i := 0; i < p.concurrency; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}

	// Periodic sweeper for orphaned pending emails
	p.wg.Add(1)
	go p.sweeper(ctx)

	slog.Info("worker started", "concurrency", p.concurrency, "buffer", cap(p.queue))
}

func (p *Processor) Enqueue(emailID string) {
	select {
	case p.queue <- emailID:
	default:
		slog.Warn("worker queue full, email will be picked up by sweeper", "emailID", emailID)
	}
}

func (p *Processor) Stop() {
	close(p.done)
	close(p.queue)
	p.wg.Wait()
	slog.Info("worker stopped")
}

func (p *Processor) worker(ctx context.Context, id int) {
	defer p.wg.Done()
	for emailID := range p.queue {
		p.processEmail(ctx, emailID)
	}
}

func (p *Processor) sweeper(ctx context.Context) {
	defer p.wg.Done()
	ticker := time.NewTicker(sweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.done:
			return
		case <-ticker.C:
			p.sweepPending(ctx)
			p.sweepRetryable(ctx)
			p.sweepAutoArchive(ctx)
			p.sweepOrphanedDigests(ctx)
			p.sweepWebhookPayloads(ctx)
		}
	}
}

// sweepWebhookPayloads enforces the retention window on stored raw webhook
// payloads. Each row holds a complete copy of an original message, so keeping
// them indefinitely steadily grows the amount of message content at rest.
func (p *Processor) sweepWebhookPayloads(ctx context.Context) {
	if p.WebhookPayloads == nil || p.PayloadRetention <= 0 {
		return
	}
	cutoff := time.Now().Add(-p.PayloadRetention)
	deleted, err := p.WebhookPayloads.DeleteOlderThan(ctx, cutoff)
	if err != nil {
		slog.Error("sweeper: failed to delete expired webhook payloads", "error", err)
		return
	}
	if deleted > 0 {
		slog.Info("sweeper: deleted expired webhook payloads", "count", deleted, "olderThan", cutoff)
	}
}

func (p *Processor) sweepPending(ctx context.Context) {
	pending, err := p.emails.ListPending(ctx, sweepBatchSize)
	if err != nil {
		slog.Error("sweeper: failed to list pending emails", "error", err)
		return
	}
	if len(pending) == 0 {
		return
	}
	for _, e := range pending {
		p.Enqueue(e.ID)
	}
	slog.Info("sweeper: re-queued orphaned emails", "found", len(pending))
}

func (p *Processor) sweepRetryable(ctx context.Context) {
	retryable, err := p.emails.ListRetryable(ctx, retryMaxAttempts, retryBatchSize)
	if err != nil {
		slog.Error("sweeper: failed to list retryable emails", "error", err)
		return
	}
	if len(retryable) == 0 {
		return
	}
	for _, e := range retryable {
		// Increment retry count and reset to pending before enqueuing
		if err := p.emails.IncrementRetryCount(ctx, e.ID); err != nil {
			slog.Error("sweeper: failed to increment retry count", "emailID", e.ID, "error", err)
			continue
		}
		if err := p.emails.SetStatus(ctx, e.ID, domain.EmailStatusPending, ""); err != nil {
			slog.Error("sweeper: failed to reset status for retry", "emailID", e.ID, "error", err)
			continue
		}
		p.Enqueue(e.ID)
	}
	slog.Info("sweeper: retrying failed emails", "found", len(retryable))
}

func (p *Processor) processEmail(ctx context.Context, emailID string) {
	ProcessEmail(ctx, ProcessEmailParams{
		EmailID:     emailID,
		Emails:      p.emails,
		Extractions: p.extractions,
		Resolver:    p.resolver,
		SenderPrefs: p.senderPrefs,
		TokenUsage:  p.tokenUsage,
		Preferences: p.preferences,
		Broker:      p.SSEBroker,
	})

	// Check token warning after processing (best-effort, non-blocking)
	if p.Mailer != nil && p.Sessions != nil && p.tokenUsage != nil && p.preferences != nil {
		em, err := p.emails.GetByIDInternal(ctx, emailID)
		if err == nil && em != nil {
			p.checkTokenWarning(ctx, em.UserID)
		}
	}
}

// ProcessEmailParams groups the dependencies for ProcessEmail to avoid a long parameter list.
type ProcessEmailParams struct {
	EmailID     string
	Emails      database.EmailRepository
	Extractions database.ExtractionRepository
	Resolver    *llm.ProviderResolver
	SenderPrefs database.SenderPreferenceRepository
	TokenUsage  database.TokenUsageRepository
	Preferences database.PreferenceRepository
	Broker      *sse.Broker
}

// ProcessEmail runs LLM extraction on a single email. It can be called
// synchronously from the webhook handler or asynchronously from the worker.
func ProcessEmail(ctx context.Context, params ProcessEmailParams) {
	emailID := params.EmailID
	emails := params.Emails
	extractions := params.Extractions
	resolver := params.Resolver
	senderPrefs := params.SenderPrefs
	tokenUsage := params.TokenUsage
	preferences := params.Preferences
	broker := params.Broker

	logger := slog.With("emailID", emailID)

	if err := emails.SetStatus(ctx, emailID, domain.EmailStatusProcessing, ""); err != nil {
		logger.Error("failed to set processing status", "error", err)
		return
	}

	em, err := emails.GetByIDInternal(ctx, emailID)
	if err != nil {
		logger.Error("failed to fetch email", "error", err)
		emails.SetStatus(ctx, emailID, domain.EmailStatusFailed, "failed to fetch email")
		return
	}

	// Check quiet hours — defer processing if user is in quiet hours
	if preferences != nil {
		pref, prefErr := preferences.Get(ctx, em.UserID)
		if prefErr == nil && pref.QuietHoursEnabled && isInQuietHours(pref) {
			logger.Info("user in quiet hours, deferring processing")
			emails.SetStatus(ctx, emailID, domain.EmailStatusPending, "")
			return
		}
	}

	// Resolve LLM provider for this user (BYOK or platform fallback)
	provider, err := resolver.ResolveForUser(ctx, em.UserID)
	if err != nil {
		logger.Error("failed to resolve LLM provider", "error", err)
		emails.SetStatus(ctx, emailID, domain.EmailStatusFailed, "failed to resolve LLM provider")
		return
	}

	// Check if sender is blocked
	if senderPrefs != nil {
		pref, _ := senderPrefs.GetByAddress(ctx, em.UserID, em.FromAddress)
		if pref != nil && pref.Status == "blocked" {
			logger.Info("sender is blocked, skipping", "from", em.FromAddress)
			emails.SetStatus(ctx, emailID, domain.EmailStatusSkipped, "sender is blocked")
			return
		}
	}

	// Check token limits (monthly and daily)
	if preferences != nil && tokenUsage != nil {
		pref, err := preferences.Get(ctx, em.UserID)
		if err == nil {
			if pref.MonthlyTokenLimit > 0 {
				used, err := tokenUsage.GetMonthlyTotal(ctx, em.UserID)
				if err == nil && used >= pref.MonthlyTokenLimit {
					logger.Warn("monthly token limit exceeded", "used", used, "limit", pref.MonthlyTokenLimit)
					emails.SetStatus(ctx, emailID, domain.EmailStatusSkipped, "monthly token limit exceeded")
					return
				}
			}
			if pref.DailyTokenLimit > 0 {
				dailyUsed, err := tokenUsage.GetDailyTotal(ctx, em.UserID)
				if err == nil && dailyUsed >= pref.DailyTokenLimit {
					logger.Warn("daily token limit exceeded", "used", dailyUsed, "limit", pref.DailyTokenLimit)
					emails.SetStatus(ctx, emailID, domain.EmailStatusPending, "daily token limit exceeded, will retry tomorrow")
					return
				}
			}
		}
	}

	// Prefer HTML body converted to markdown — preserves headings, links,
	// and structure that the LLM can use for better extraction.
	content := ""
	if em.HTMLBody != "" {
		content = emailpkg.HTMLToMarkdown(em.HTMLBody)
	}
	if content == "" {
		content = em.TextBody
	}
	if content == "" {
		logger.Warn("email has no content, skipping")
		emails.SetStatus(ctx, emailID, domain.EmailStatusFailed, "email has no content")
		return
	}

	// Triage: cheap LLM call to decide if this email is worth extracting.
	contentPreview := content
	if len(contentPreview) > triagePreviewMaxLen {
		contentPreview = contentPreview[:triagePreviewMaxLen]
	}
	triageResult, triageUsage, triageErr := provider.TriageEmail(ctx, em.Subject, em.FromAddress, contentPreview)
	// Always record triage tokens when available, even if parsing failed
	recordTokenUsage(ctx, tokenUsage, em.UserID, "triage", provider, triageUsage)
	if triageErr != nil {
		logger.Warn("triage failed, proceeding to extraction", "error", triageErr)
	}
	if triageErr == nil && !triageResult.Extract {
		logger.Info("triage skipped email", "reason", triageResult.Reason)
		emails.SetStatus(ctx, emailID, domain.EmailStatusSkipped, triageResult.Reason)
		return
	}

	// Retry loop for extraction with exponential backoff
	var result *llm.ExtractionResult
	var usage *llm.Usage
	for attempt := 1; attempt <= extractMaxRetries; attempt++ {
		result, usage, err = provider.ExtractEmail(ctx, em.Subject, content, em.FromAddress)
		if err == nil {
			break
		}
		if attempt < extractMaxRetries {
			delay := extractBaseDelay * time.Duration(1<<(attempt-1)) // 2s, 4s
			logger.Warn("LLM extraction failed, retrying", "attempt", attempt, "delay", delay, "error", err)
			select {
			case <-ctx.Done():
				logger.Error("context cancelled during extraction retry", "error", ctx.Err())
				emails.SetStatus(ctx, emailID, domain.EmailStatusFailed, "context cancelled")
				return
			case <-time.After(delay):
			}
		}
	}
	if err != nil {
		logger.Error("LLM extraction failed after retries", "attempts", extractMaxRetries, "error", err)
		emails.SetStatus(ctx, emailID, domain.EmailStatusFailed, "extraction failed")
		return
	}

	// Apply sender category override if user has set one
	sourceCategory := result.SourceCategory
	if senderPrefs != nil {
		pref, _ := senderPrefs.GetByAddress(ctx, em.UserID, em.FromAddress)
		if pref != nil && pref.Category != "" {
			sourceCategory = pref.Category
		}
	}

	now := time.Now()
	extraction := &domain.Extraction{
		ID:             uuid.New().String(),
		EmailID:        emailID,
		Summary:        result.Summary,
		KeyPoints:      result.KeyPoints,
		Topics:         result.Topics,
		Links:          result.Links,
		Sentiment:      result.Sentiment,
		SourceCategory: sourceCategory,
		Provider:       provider.Name(),
		Model:          provider.Model(),
		TokensUsed:     usage.TotalTokens,
		ProcessedAt:    now,
		CreatedAt:      now,
	}

	if err := extractions.Create(ctx, extraction); err != nil {
		logger.Error("failed to store extraction", "error", err)
		emails.SetStatus(ctx, emailID, domain.EmailStatusFailed, "failed to store extraction")
		return
	}

	if err := emails.SetStatus(ctx, emailID, domain.EmailStatusProcessed, ""); err != nil {
		logger.Error("failed to set processed status", "error", err)
		return
	}

	recordTokenUsage(ctx, tokenUsage, em.UserID, "extract", provider, usage)

	if broker != nil {
		broker.Publish(em.UserID, sse.Event{
			Type:    "email_processed",
			EmailID: emailID,
		})
	}

	logger.Info("email processed",
		"provider", provider.Name(),
		"tokens", usage.TotalTokens,
	)
}

func recordTokenUsage(ctx context.Context, repo database.TokenUsageRepository, userID, operation string, provider llm.Provider, usage *llm.Usage) {
	if repo == nil || usage == nil {
		return
	}
	tu := &domain.TokenUsage{
		ID:           uuid.New().String(),
		UserID:       userID,
		Operation:    operation,
		Provider:     provider.Name(),
		Model:        provider.Model(),
		InputTokens:  usage.InputTokens,
		OutputTokens: usage.OutputTokens,
		TotalTokens:  usage.TotalTokens,
		CreatedAt:    time.Now(),
	}
	if err := repo.Create(ctx, tu); err != nil {
		slog.Warn("failed to record token usage", "operation", operation, "error", err)
	}
}

func (p *Processor) checkTokenWarning(ctx context.Context, userID string) {
	pref, err := p.preferences.Get(ctx, userID)
	if err != nil || pref.MonthlyTokenLimit <= 0 {
		return
	}

	// Only send warning once per month
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	if pref.TokenWarningSentAt != nil && pref.TokenWarningSentAt.After(monthStart) {
		return
	}

	monthlyUsed, err := p.tokenUsage.GetMonthlyTotal(ctx, userID)
	if err != nil {
		return
	}

	usagePercent := monthlyUsed * 100 / pref.MonthlyTokenLimit
	if usagePercent < 80 {
		return
	}

	userEmail, err := p.Sessions.GetUserEmail(ctx, userID)
	if err != nil || userEmail == "" {
		return
	}

	if err := p.Mailer.SendTokenWarning(ctx, userEmail, usagePercent, monthlyUsed, pref.MonthlyTokenLimit); err != nil {
		slog.Warn("failed to send token warning email", "user_id", userID, "error", err)
		return
	}

	if err := p.preferences.SetTokenWarningSent(ctx, userID); err != nil {
		slog.Warn("failed to update token warning timestamp", "user_id", userID, "error", err)
	}

	slog.Info("sent token warning email", "user_id", userID, "usage_percent", usagePercent)
}

func isInQuietHours(pref *domain.UserPreference) bool {
	loc, err := time.LoadLocation(pref.DigestTimezone)
	if err != nil {
		loc = time.UTC
	}
	hour := time.Now().In(loc).Hour()
	start := pref.QuietHoursStart
	end := pref.QuietHoursEnd
	if start <= end {
		// e.g., start=1, end=6 → quiet from 1am to 6am
		return hour >= start && hour < end
	}
	// Wraps midnight: e.g., start=22, end=7 → quiet from 10pm to 7am
	return hour >= start || hour < end
}

func (p *Processor) sweepAutoArchive(ctx context.Context) {
	if p.AutoArchiveRules == nil {
		return
	}
	candidates, err := p.AutoArchiveRules.ListEmailsToArchive(ctx, 200)
	if err != nil {
		slog.Error("auto-archive sweep failed", "error", err)
		return
	}
	if len(candidates) == 0 {
		return
	}
	// Group by user
	byUser := map[string][]string{}
	for _, c := range candidates {
		byUser[c.UserID] = append(byUser[c.UserID], c.EmailID)
	}
	archived := 0
	isArchived := true
	for userID, ids := range byUser {
		if err := p.emails.BatchUpdateState(ctx, userID, ids, domain.EmailState{IsArchived: &isArchived}); err != nil {
			slog.Error("auto-archive batch update failed", "userID", userID, "error", err)
			continue
		}
		archived += len(ids)
	}
	if archived > 0 {
		slog.Info("auto-archive sweep complete", "archived", archived)
	}
}

func (p *Processor) sweepOrphanedDigests(ctx context.Context) {
	if p.Digests == nil {
		return
	}
	deleted, err := p.Digests.DeleteOrphaned(ctx)
	if err != nil {
		slog.Error("orphaned digest sweep failed", "error", err)
		return
	}
	if deleted > 0 {
		slog.Info("orphaned digest sweep complete", "deleted", deleted)
	}
}
