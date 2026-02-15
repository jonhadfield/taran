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
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/digest"
	"github.com/hadfielj/taran/backend/internal/handler"
	"github.com/hadfielj/taran/backend/internal/llm"
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

	// LLM Provider
	provider, err := newLLMProvider(cfg)
	if err != nil {
		slog.Error("failed to create LLM provider", "error", err)
		os.Exit(1)
	}
	slog.Info("LLM provider configured", "provider", provider.Name(), "model", provider.Model())

	// Background worker
	proc := worker.NewProcessor(100, 2, emailRepo, extractionRepo, provider)
	proc.Start(ctx)

	// Digest scheduler
	gen := &digest.Generator{
		Emails:      emailRepo,
		Extractions: extractionRepo,
		Digests:     digestRepo,
		Accounts:    accountRepo,
		Provider:    provider,
	}
	sched, err := digest.NewScheduler(cfg.Digest.Cron, cfg.Digest.Timezone, gen, emailRepo)
	if err != nil {
		slog.Error("failed to create digest scheduler", "error", err)
		os.Exit(1)
	}
	sched.Start()

	// Handlers
	webhookHandler := &handler.WebhookHandler{
		Accounts:    accountRepo,
		Emails:      emailRepo,
		Extractions: extractionRepo,
		Provider:    provider,
	}
	emailHandler := &handler.EmailHandler{
		Emails:      emailRepo,
		Extractions: extractionRepo,
	}
	digestHandler := &handler.DigestHandler{
		Digests:   digestRepo,
		Generator: gen,
	}
	accountHandler := &handler.AccountHandler{
		Accounts:    accountRepo,
		EmailDomain: cfg.Email.Domain,
	}
	sessionAuth := &auth.SessionAuth{
		Sessions:    sessionRepo,
		AdminEmails: cfg.AdminEmails,
	}

	// HTTP server
	mux := server.NewRouter(server.RouterDeps{
		WebhookSecret:  cfg.Webhook.Secret,
		APIKey:         cfg.Server.APIKey,
		WebhookHandler: webhookHandler,
		EmailHandler:   emailHandler,
		DigestHandler:  digestHandler,
		AccountHandler: accountHandler,
		SessionAuth:    sessionAuth,
	})
	cors := server.CORSMiddleware(cfg.Server.AllowedOrigins)
	httpHandler := server.RecoveryMiddleware(cors(server.LoggingMiddleware(mux)))

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
	sched.Stop()

	slog.Info("shutdown complete")
}

func newLLMProvider(cfg *config.Config) (llm.Provider, error) {
	switch cfg.LLM.Provider {
	case "anthropic":
		return llm.NewAnthropicProvider(cfg.LLM.AnthropicKey, cfg.LLM.AnthropicModel), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.LLM.Provider)
	}
}
