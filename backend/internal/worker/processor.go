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

type Processor struct {
	queue       chan string
	emails      database.EmailRepository
	extractions database.ExtractionRepository
	provider    llm.Provider
	senderPrefs database.SenderPreferenceRepository
	wg          sync.WaitGroup
	concurrency int
}

func NewProcessor(
	bufferSize, concurrency int,
	emails database.EmailRepository,
	extractions database.ExtractionRepository,
	provider llm.Provider,
	senderPrefs database.SenderPreferenceRepository,
) *Processor {
	return &Processor{
		queue:       make(chan string, bufferSize),
		emails:      emails,
		extractions: extractions,
		provider:    provider,
		senderPrefs: senderPrefs,
		concurrency: concurrency,
	}
}

func (p *Processor) Start(ctx context.Context) {
	pending, err := p.emails.ListPending(ctx, 100)
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

	slog.Info("worker started", "concurrency", p.concurrency, "buffer", cap(p.queue))
}

func (p *Processor) Enqueue(emailID string) {
	select {
	case p.queue <- emailID:
	default:
		slog.Warn("worker queue full, email will be picked up on restart", "emailID", emailID)
	}
}

func (p *Processor) Stop() {
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

func (p *Processor) processEmail(ctx context.Context, emailID string) {
	ProcessEmail(ctx, emailID, p.emails, p.extractions, p.provider, p.senderPrefs)
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
) {
	logger := slog.With("emailID", emailID)

	if err := emails.SetStatus(ctx, emailID, domain.EmailStatusProcessing); err != nil {
		logger.Error("failed to set processing status", "error", err)
		return
	}

	em, err := emails.GetByIDInternal(ctx, emailID)
	if err != nil {
		logger.Error("failed to fetch email", "error", err)
		emails.SetStatus(ctx, emailID, domain.EmailStatusFailed)
		return
	}

	// Check if sender is blocked
	if senderPrefs != nil {
		pref, _ := senderPrefs.GetByAddress(ctx, em.UserID, em.FromAddress)
		if pref != nil && pref.Status == "blocked" {
			logger.Info("sender is blocked, skipping", "from", em.FromAddress)
			emails.SetStatus(ctx, emailID, domain.EmailStatusSkipped)
			return
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
		emails.SetStatus(ctx, emailID, domain.EmailStatusFailed)
		return
	}

	// Triage: cheap LLM call to decide if this email is worth extracting.
	contentPreview := content
	if len(contentPreview) > 500 {
		contentPreview = contentPreview[:500]
	}
	triageResult, _, triageErr := provider.TriageEmail(ctx, em.Subject, em.FromAddress, contentPreview)
	if triageErr != nil {
		logger.Warn("triage failed, proceeding to extraction", "error", triageErr)
	} else if !triageResult.Extract {
		logger.Info("triage skipped email", "reason", triageResult.Reason)
		emails.SetStatus(ctx, emailID, domain.EmailStatusSkipped)
		return
	}

	result, usage, err := provider.ExtractEmail(ctx, em.Subject, content, em.FromAddress)
	if err != nil {
		logger.Error("LLM extraction failed", "error", err)
		emails.SetStatus(ctx, emailID, domain.EmailStatusFailed)
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
		emails.SetStatus(ctx, emailID, domain.EmailStatusFailed)
		return
	}

	if err := emails.SetStatus(ctx, emailID, domain.EmailStatusProcessed); err != nil {
		logger.Error("failed to set processed status", "error", err)
		return
	}

	logger.Info("email processed",
		"provider", provider.Name(),
		"tokens", usage.TotalTokens,
	)
}
