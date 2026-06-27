# MailBrief Backend

Go API server handling email ingestion, AI-powered extraction, digest generation, and the REST API.

## Tech Stack

- **Go 1.24** with stdlib `net/http` routing (no framework)
- **pgx/v5** for PostgreSQL (no ORM)
- **enmime** for RFC 5322 email parsing
- **Anthropic SDK v1.22** + **OpenAI Go SDK v3** for LLM integration
- **Resend** for outbound email delivery
- **log/slog** for structured logging

## Architecture

```
cmd/taran/main.go          Entry point: config, DB, migrations, DI wiring, graceful shutdown
│
├── internal/server/        HTTP server, router, middleware chain
├── internal/auth/          Session validation (Better Auth), webhook auth, API key auth
├── internal/handler/       HTTP handlers (one file per domain)
├── internal/email/         RFC 5322 parser + HTML sanitiser
├── internal/llm/           LLM provider interface + implementations
├── internal/worker/        Background email processing (goroutine pool)
├── internal/digest/        Digest generator + cron scheduler
├── internal/database/      Repository interfaces + pgx implementations
├── internal/domain/        Pure domain types (no dependencies)
├── internal/config/        Environment-based configuration
│
└── migrations/             SQL migration files (embedded, run at startup)
```

## HTTP Layer

### Middleware Chain

Applied in order to every request:

1. **Recovery** -- catches panics, returns 500
2. **Logging** -- logs method, path, status, duration via `slog`
3. **CORS** -- validates origin, handles preflight
4. **Rate Limiting** -- split limits: API (10 req/s, 30 burst) vs webhooks (50 req/s, 100 burst)
5. **Security Headers** -- CSP, X-Content-Type-Options, X-Frame-Options, Referrer-Policy

### Authentication

Three auth schemes, applied per-route:

| Scheme | Header/Cookie | Used By |
|--------|--------------|---------|
| **Session Auth** | `Authorization: Bearer {token}` or `auth_token` cookie | All `/api/*` routes |
| **Webhook Auth** | `X-Webhook-Secret` | `POST /webhook/email`, `POST /cron/digests` |
| **API Key Auth** | `X-API-Key` | All `/api/*` routes (additional layer) |

Session auth validates tokens against Better Auth's `session` table directly (camelCase column names: `"userId"`, `"expiresAt"`).

### Key Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/webhook/email` | Receive raw email from Cloudflare Worker |
| `POST` | `/cron/digests` | Trigger digest generation (Cloud Scheduler) |
| `GET` | `/api/emails` | List emails (paginated, filterable) |
| `GET` | `/api/emails/{id}` | Email detail with extraction |
| `GET/POST` | `/api/digests` | List or generate digests |
| `GET` | `/api/digests/{id}` | Digest detail with items |
| `GET` | `/api/dashboard` | Dashboard overview (stats, recent data) |
| `GET` | `/api/stats/*` | Analytics (volume, categories, heatmap) |
| `GET/PATCH` | `/api/preferences` | User preferences |
| `GET/PATCH` | `/api/senders/*` | Sender management |
| `GET/POST/DELETE` | `/api/labels/*` | Label CRUD + email assignment |
| `GET/POST/DELETE` | `/api/saved-searches` | Filter presets |
| `GET/PUT/DELETE` | `/api/auto-archive-rules` | Auto-archive rules |
| `GET` | `/api/public/digests/{token}` | Public shared digest |
| `POST` | `/api/admin/*` | Admin operations |

Full spec available at `/api/openapi.yaml` with interactive docs at `/docs`.

## Email Processing Pipeline

When an email arrives at `POST /webhook/email`:

```
1. Parse RFC 5322 (enmime) → extract headers, bodies, attachments, threading info
2. Resolve thread ID via In-Reply-To / References headers
3. Create Email record (status: "pending")
4. Store webhook payload (dead letter tracking)
5. Enqueue to background worker
```

The background worker then processes each email:

```
1. Check quiet hours → defer if in user's quiet window
2. Resolve LLM provider → BYOK key or platform fallback
3. Check sender blocked → skip if blocked
4. Check token limits → skip if monthly/daily limit hit
5. Convert HTML → Markdown for LLM input
6. Triage (cheap LLM call, 30s timeout) → decide if worth extracting
7. Extract (expensive LLM call, 90s timeout, 3 retries) → structured data
8. Apply sender category override
9. Store extraction in DB
10. Check 80% token warning → email user if threshold hit
11. Apply auto-archive rules
```

**Worker configuration:** 2 concurrent workers, 100-item queue, 5-minute sweep interval for retries (exponential backoff, max 5 retries).

## LLM Integration

Pluggable provider interface with three operations:

| Operation | Timeout | Purpose |
|-----------|---------|---------|
| `TriageEmail` | 30s | Cheap call to decide if an email is worth extracting |
| `ExtractEmail` | 90s | Extract summary, key points, topics, links, sentiment |
| `GenerateDigest` | 90s | Summarise a batch of extractions into a digest |

### Providers

| Provider | Implementation | Notes |
|----------|---------------|-------|
| **Anthropic** | `anthropic-sdk-go` v1.22 | Primary provider |
| **OpenAI** | `openai-go` v3 | Secondary/fallback |
| **Ollama** | HTTP API | Local development |
| **Fallback** | Wraps primary + secondary | Auto-retries on transient errors |

### BYOK (Bring Your Own Key)

Users can store their own API keys (encrypted with AES-256-GCM). The `ProviderResolver` checks for a user's active key before falling back to the platform provider.

## Digest Generation

Triggered by Cloud Scheduler hitting `POST /cron/digests`:

1. **Mutex lock** prevents duplicate runs from concurrent triggers
2. **Retry orphaned digests** generated >10 min ago but never sent
3. For each user with pending extractions:
   - Filter extractions (exclude muted/blocked senders, apply feedback, apply keyword and category filters)
   - Sort by priority (feedback-boosted items first)
   - Generate digest via LLM
   - Create digest record with share token
   - Send via Resend if email delivery enabled

**Duplicate prevention:** DB unique constraint `uq_digest_user_period` + `sync.Mutex` in-process.

## Database

PostgreSQL with `pgx/v5`. Repository pattern -- one interface per entity, implementations use raw SQL.

### Migration System

- SQL files in `migrations/` embedded into the binary via `//go:embed`
- Applied at startup in order, transactional, fatal on failure
- Version tracked in `schema_migrations` table

### Key Tables

| Table | Owner | Purpose |
|-------|-------|---------|
| `user`, `session`, `account` | Better Auth (Next.js) | Authentication |
| `email_account` | Go backend | Managed inboxes |
| `email` | Go backend | Stored emails with status tracking |
| `extraction` | Go backend | LLM-extracted structured data (JSONB arrays) |
| `digest`, `digest_item` | Go backend | Generated digests |
| `user_preference` | Go backend | Per-user settings |
| `sender_preference` | Go backend | Per-sender status and category overrides |
| `token_usage` | Go backend | LLM token consumption tracking |
| `label`, `email_label` | Go backend | User-defined labels |
| `webhook_payload` | Go backend | Dead letter queue |

### Encryption at Rest

Email HTML and text bodies are encrypted with AES-256-GCM in the repository layer when an encryption key is configured. BYOK API keys are also stored encrypted.

## Development

```bash
# Build
go build -o taran ./cmd/taran/

# Cross-compile for deployment
GOOS=linux GOARCH=amd64 go build -o taran-linux-amd64 ./cmd/taran/

# Run (requires PostgreSQL and .env)
./taran
```

Or via the root Makefile:

```bash
make db              # Start PostgreSQL (Docker)
make build-backend   # Build binary
make start-backend   # Build + start in background
make stop-backend    # Stop
```

## Configuration

All configuration via environment variables (loaded from `.env` via `godotenv`):

| Variable | Purpose |
|----------|---------|
| `DATABASE_URL` | PostgreSQL connection string |
| `WEBHOOK_SECRET` | Shared secret for Cloudflare Worker auth |
| `LLM_PROVIDER` | Primary provider (`anthropic`, `openai`, `ollama`) |
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `OPENAI_API_KEY` | OpenAI API key (enables fallback if both set) |
| `RESEND_API_KEY` | Resend API key for outbound email |
| `API_KEY` | API key for frontend-to-backend auth |
| `ALLOWED_ORIGINS` | CORS allowed origins |
| `ENCRYPTION_KEY` | AES-256 key for email encryption at rest |
| `ADMIN_EMAILS` | Comma-separated admin email addresses |
