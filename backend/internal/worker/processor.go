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
)

const (
	extractMaxRetries = 3
	extractBaseDelay  = 2 * time.Second
	sweepInterval     = 5 * time.Minute
	sweepBatchSize    = 100
	startupRequeueMax = 500
	retryMaxAttempts  = 5 // max auto-retries for failed emails (backoff: 5m, 20m, 80m, 320m, 1280m)
	retryBatchSize    = 20
)

type Processor struct {
	queue       chan string
	done        chan struct{}
	emails      database.EmailRepository
	extractions database.ExtractionRepository
	provider    llm.Provider
	senderPrefs database.SenderPreferenceRepository
	tokenUsage  database.TokenUsageRepository
	preferences database.PreferenceRepository
	wg          sync.WaitGroup
	concurrency int
}

func NewProcessor(
	bufferSize, concurrency int,
	emails database.EmailRepository,
	extractions database.ExtractionRepository,
	provider llm.Provider,
	senderPrefs database.SenderPreferenceRepository,
	tokenUsage database.TokenUsageRepository,
	preferences database.PreferenceRepository,
) *Processor {
	return &Processor{
		queue:       make(chan string, bufferSize),
		done:        make(chan struct{}),
		emails:      emails,
		extractions: extractions,
		provider:    provider,
		senderPrefs: senderPrefs,
		tokenUsage:  tokenUsage,
		preferences: preferences,
		concurrency: concurrency,
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
		}
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
	enqueued := 0
	for _, e := range pending {
		select {
		case p.queue <- e.ID:
			enqueued++
		default:
			break
		}
	}
	slog.Info("sweeper: re-queued orphaned emails", "found", len(pending), "enqueued", enqueued)
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
	enqueued := 0
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
		select {
		case p.queue <- e.ID:
			enqueued++
		default:
			break
		}
	}
	slog.Info("sweeper: retrying failed emails", "found", len(retryable), "enqueued", enqueued)
}

func (p *Processor) processEmail(ctx context.Context, emailID string) {
	ProcessEmail(ctx, emailID, p.emails, p.extractions, p.provider, p.senderPrefs, p.tokenUsage, p.preferences)
}

// ProcessEmail runs LLM extraction on a single email. It can be called
// synchronously from the webhook handler or asynchronously from the worker.
func ProcessEmail(
	ctx context.Context,
	emailID string,
	emails database.EmailRepository,
	extractions database.ExtractionRepository,
	provider llm.Provider,
	senderPrefs database.SenderPreferenceRepository,
	tokenUsage database.TokenUsageRepository,
	preferences database.PreferenceRepository,
) {
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

	// Check if sender is blocked
	if senderPrefs != nil {
		pref, _ := senderPrefs.GetByAddress(ctx, em.UserID, em.FromAddress)
		if pref != nil && pref.Status == "blocked" {
			logger.Info("sender is blocked, skipping", "from", em.FromAddress)
			emails.SetStatus(ctx, emailID, domain.EmailStatusSkipped, "sender is blocked")
			return
		}
	}

	// Check monthly token limit
	if preferences != nil && tokenUsage != nil {
		pref, err := preferences.Get(ctx, em.UserID)
		if err == nil && pref.MonthlyTokenLimit > 0 {
			used, err := tokenUsage.GetMonthlyTotal(ctx, em.UserID)
			if err == nil && used >= pref.MonthlyTokenLimit {
				logger.Warn("monthly token limit exceeded", "used", used, "limit", pref.MonthlyTokenLimit)
				emails.SetStatus(ctx, emailID, domain.EmailStatusSkipped, "monthly token limit exceeded")
				return
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
	if len(contentPreview) > 500 {
		contentPreview = contentPreview[:500]
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

	now := time.Now()
	extraction := &domain.Extraction{
		ID:             uuid.New().String(),
		EmailID:        emailID,
		Summary:        result.Summary,
		KeyPoints:      result.KeyPoints,
		Topics:         result.Topics,
		Links:          result.Links,
		ActionItems:    result.ActionItems,
		Sentiment:      result.Sentiment,
		SourceCategory: result.SourceCategory,
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
