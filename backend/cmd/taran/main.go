package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"log/syslog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/config"
	"github.com/hadfielj/taran/backend/internal/crypto"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/digest"
	"github.com/hadfielj/taran/backend/internal/handler"
	"github.com/hadfielj/taran/backend/internal/llm"
	"github.com/hadfielj/taran/backend/internal/mailer"
	"github.com/hadfielj/taran/backend/internal/server"
	"github.com/hadfielj/taran/backend/internal/worker"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	var logWriter io.Writer = os.Stdout
	if os.Getenv("TARAN_LOG_SYSLOG") == "true" {
		sw, err := syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, "taran")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect to syslog: %v\n", err)
			os.Exit(1)
		}
		logWriter = sw
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(logWriter, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := database.NewPool(ctx, cfg.DB.URL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := database.RunMigrations(ctx, pool); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Repositories
	accountRepo := database.NewAccountRepo(pool)
	emailRepo := database.NewEmailRepo(pool)
	extractionRepo := database.NewExtractionRepo(pool)
	digestRepo := database.NewDigestRepo(pool)
	sessionRepo := database.NewSessionRepo(pool)
	appSettingRepo := database.NewAppSettingRepo(pool)
	preferenceRepo := database.NewPreferenceRepo(pool)
	preferenceRepo.AppSettings = appSettingRepo
	senderPrefRepo := database.NewSenderPreferenceRepo(pool)
	feedbackRepo := database.NewFeedbackRepo(pool)
	inviteRepo := database.NewInviteRepo(pool)
	waitlistRepo := database.NewWaitlistRepo(pool)
	digestFeedbackRepo := database.NewDigestFeedbackRepo(pool)
	attachmentRepo := database.NewAttachmentRepo(pool)
	tokenUsageRepo := database.NewTokenUsageRepo(pool)

	// LLM Provider (with optional fallback)
	provider, err := newLLMProvider(cfg)
	if err != nil {
		slog.Error("failed to create LLM provider", "error", err)
		os.Exit(1)
	}
	slog.Info("LLM provider configured", "provider", provider.Name(), "model", provider.Model())

	// If both API keys are set, wrap in a fallback provider
	if secondary := newSecondaryProvider(cfg); secondary != nil {
		slog.Info("LLM fallback enabled", "primary", provider.Name(), "secondary", secondary.Name())
		provider = llm.NewFallbackProvider(provider, secondary)
	} else if cfg.LLM.AutoSelectedOverAnthropic {
		slog.Warn("both Anthropic and OpenAI API keys set, using OpenAI", "model", provider.Model())
	}

	// BYOK encryption (optional)
	userLLMKeyRepo := database.NewUserLLMKeyRepo(pool)
	var encryptor *crypto.Encryptor
	if cfg.LLM.EncryptionKey != "" {
		encryptor, err = crypto.NewEncryptor(cfg.LLM.EncryptionKey)
		if err != nil {
			slog.Error("failed to create encryptor", "error", err)
			os.Exit(1)
		}
		slog.Info("BYOK encryption enabled")
	}

	// Provider resolver (BYOK → platform fallback)
	resolver := llm.NewProviderResolver(provider, userLLMKeyRepo, encryptor, &cfg.LLM)

	// Background worker
	proc := worker.NewProcessor(100, 2, emailRepo, extractionRepo, resolver, senderPrefRepo, tokenUsageRepo, preferenceRepo)

	// Mailer (optional — disabled if no Resend API key)
	var m mailer.Mailer
	if cfg.Email.ResendAPIKey != "" {
		m = mailer.NewResendMailer(cfg.Email.ResendAPIKey, cfg.Email.FromAddress)
		slog.Info("digest email delivery enabled via Resend")
	}

	// Wire token warning deps into processor before starting
	proc.Mailer = m
	proc.Sessions = sessionRepo
	proc.Start(ctx)

	// Digest scheduler
	gen := &digest.Generator{
		Emails:      emailRepo,
		Extractions: extractionRepo,
		Digests:     digestRepo,
		Accounts:    accountRepo,
		Resolver:    resolver,
		SenderPrefs: senderPrefRepo,
		Feedback:    feedbackRepo,
		Preferences: preferenceRepo,
		TokenUsage:  tokenUsageRepo,
	}
	sched := digest.NewScheduler(gen, emailRepo, digestRepo, preferenceRepo, sessionRepo, m, cfg.Server.BaseURL, cfg.Server.UnsubscribeSecret)

	// Webhook payload repo (for dead letter queue / replay)
	webhookPayloadRepo := database.NewWebhookPayloadRepo(pool)

	// Label repo
	labelRepo := database.NewLabelRepo(pool)

	// Saved search repo
	savedSearchRepo := database.NewSavedSearchRepo(pool)

	// Auto-archive rule repo
	autoArchiveRepo := database.NewAutoArchiveRuleRepo(pool)
	proc.AutoArchiveRules = autoArchiveRepo

	// Handlers
	webhookHandler := &handler.WebhookHandler{
		Accounts:        accountRepo,
		Emails:          emailRepo,
		Extractions:     extractionRepo,
		Attachments:     attachmentRepo,
		WebhookPayloads: webhookPayloadRepo,
		Resolver:        resolver,
		SenderPrefs:     senderPrefRepo,
		TokenUsage:      tokenUsageRepo,
		Preferences:     preferenceRepo,
	}
	emailHandler := &handler.EmailHandler{
		Emails:      emailRepo,
		Extractions: extractionRepo,
		Attachments: attachmentRepo,
		Processor:   proc,
		SenderPrefs: senderPrefRepo,
	}
	digestHandler := &handler.DigestHandler{
		Digests:           digestRepo,
		DigestFeedback:    digestFeedbackRepo,
		Generator:         gen,
		Preferences:       preferenceRepo,
		Mailer:            m,
		Sessions:          sessionRepo,
		BaseURL:           cfg.Server.BaseURL,
		UnsubscribeSecret: cfg.Server.UnsubscribeSecret,
	}
	accountHandler := &handler.AccountHandler{
		Accounts:    accountRepo,
		EmailDomain: cfg.Email.Domain,
	}
	senderHandler := &handler.SenderHandler{
		Emails:      emailRepo,
		SenderPrefs: senderPrefRepo,
		Feedback:    feedbackRepo,
	}
	statsHandler := &handler.StatsHandler{
		Emails: emailRepo,
	}
	statsHistoryHandler := &handler.StatsHistoryHandler{
		Emails: emailRepo,
	}
	preferenceHandler := &handler.PreferenceHandler{
		Preferences:       preferenceRepo,
		UnsubscribeSecret: cfg.Server.UnsubscribeSecret,
	}
	topicHandler := &handler.TopicHandler{
		Extractions: extractionRepo,
	}
	feedbackHandler := &handler.FeedbackHandler{
		Feedback: feedbackRepo,
		Emails:   emailRepo,
	}
	dashboardHandler := &handler.DashboardHandler{
		Emails:      emailRepo,
		Digests:     digestRepo,
		Extractions: extractionRepo,
	}
	usageHandler := &handler.UsageHandler{
		TokenUsage:  tokenUsageRepo,
		Preferences: preferenceRepo,
	}
	adminStatsHandler := &handler.AdminStatsHandler{
		Pool:        pool,
		LLMProvider: provider.Name(),
		LLMModel:    provider.Model(),
		TokenUsage:  tokenUsageRepo,
		Preferences: preferenceRepo,
		AppSettings: appSettingRepo,
	}
	inviteHandler := &handler.InviteHandler{
		Invites:     inviteRepo,
		AdminEmails: cfg.AdminEmails,
		Mailer:      m,
	}
	waitlistHandler := &handler.WaitlistHandler{
		Waitlist: waitlistRepo,
		Invites:  inviteRepo,
		Mailer:   m,
	}
	exportHandler := &handler.ExportHandler{
		Emails:  emailRepo,
		Digests: digestRepo,
	}
	llmKeyHandler := &handler.LLMKeyHandler{
		Keys:      userLLMKeyRepo,
		Encryptor: encryptor,
	}
	adminWebhookHandler := &handler.AdminWebhookHandler{
		Emails:          emailRepo,
		Extractions:     extractionRepo,
		WebhookPayloads: webhookPayloadRepo,
		Accounts:        accountRepo,
		Resolver:        resolver,
		SenderPrefs:     senderPrefRepo,
		TokenUsage:      tokenUsageRepo,
		Preferences:     preferenceRepo,
		Processor:       proc,
	}
	labelHandler := &handler.LabelHandler{
		Labels: labelRepo,
	}
	autoArchiveHandler := &handler.AutoArchiveHandler{
		Rules: autoArchiveRepo,
	}
	savedSearchHandler := &handler.SavedSearchHandler{
		SavedSearches: savedSearchRepo,
	}
	cronHandler := &handler.CronHandler{
		Scheduler: sched,
	}
	sessionAuth := &auth.SessionAuth{
		Sessions:    sessionRepo,
		AdminEmails: cfg.AdminEmails,
	}

	// HTTP server
	mux := server.NewRouter(server.RouterDeps{
		Pool:               pool,
		WebhookSecret:      cfg.Webhook.Secret,
		APIKey:             cfg.Server.APIKey,
		ExportHandler:      exportHandler,
		WebhookHandler:     webhookHandler,
		EmailHandler:       emailHandler,
		DigestHandler:      digestHandler,
		AccountHandler:     accountHandler,
		PreferenceHandler:  preferenceHandler,
		SenderHandler:      senderHandler,
		StatsHandler:       statsHandler,
		StatsHistoryHandler: statsHistoryHandler,
		TopicHandler:       topicHandler,
		FeedbackHandler:    feedbackHandler,
		UsageHandler:       usageHandler,
		DashboardHandler:   dashboardHandler,
		InviteHandler:      inviteHandler,
		WaitlistHandler:    waitlistHandler,
		AdminStatsHandler:  adminStatsHandler,
		SessionAuth:        sessionAuth,
		LLMKeyHandler:        llmKeyHandler,
		CronHandler:          cronHandler,
		AdminWebhookHandler:  adminWebhookHandler,
		AutoArchiveHandler:   autoArchiveHandler,
		LabelHandler:         labelHandler,
		SavedSearchHandler:   savedSearchHandler,
	})
	cors := server.CORSMiddleware(cfg.Server.AllowedOrigins)
	limiter := server.NewRateLimiter(10, 30) // 10 req/s sustained, 30 burst
	httpHandler := server.RecoveryMiddleware(server.SecurityHeaders(cors(limiter.Middleware(server.LoggingMiddleware(mux)))))

	var tlsCfg *server.TLSConfig
	if cfg.Server.TLSDomain != "" {
		tlsCfg = &server.TLSConfig{
			Domain:  cfg.Server.TLSDomain,
			CertDir: cfg.Server.TLSCertDir,
		}
	}
	srv := server.New(cfg.Addr(), httpHandler, tlsCfg)

	go func() {
		if err := srv.Start(); err != nil {
			slog.Error("server error", "error", err)
			cancel()
		}
	}()

	if cfg.Server.TLSDomain != "" {
		slog.Info("taran is running", "https", ":443", "http", ":80", "domain", cfg.Server.TLSDomain)
	} else {
		slog.Info("taran is running", "addr", cfg.Addr())
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}
	proc.Stop()

	slog.Info("shutdown complete")
}

func newLLMProvider(cfg *config.Config) (llm.Provider, error) {
	switch cfg.LLM.Provider {
	case "anthropic":
		return llm.NewAnthropicProvider(cfg.LLM.AnthropicKey, cfg.LLM.AnthropicModel), nil
	case "openai":
		return llm.NewOpenAIProvider(cfg.LLM.OpenAIKey, cfg.LLM.OpenAIModel), nil
	case "ollama":
		return llm.NewOllamaProvider(cfg.LLM.OllamaURL, cfg.LLM.OllamaModel), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.LLM.Provider)
	}
}

// newSecondaryProvider returns a secondary LLM provider for fallback, or nil
// if only one provider key is configured. Ollama primary never gets a fallback.
func newSecondaryProvider(cfg *config.Config) llm.Provider {
	switch cfg.LLM.Provider {
	case "anthropic":
		if cfg.LLM.OpenAIKey != "" {
			return llm.NewOpenAIProvider(cfg.LLM.OpenAIKey, cfg.LLM.OpenAIModel)
		}
	case "openai":
		if cfg.LLM.AnthropicKey != "" {
			return llm.NewAnthropicProvider(cfg.LLM.AnthropicKey, cfg.LLM.AnthropicModel)
		}
	}
	return nil
}
