# MailBrief

**Your newsletters, summarised.** MailBrief is an AI-powered email digest dashboard. Sign up, get a managed inbox, forward your newsletters, and receive daily AI-generated summaries — key points, action items, topics, and sentiment — without reading every email.

**Live at [mailbrief.io](https://mailbrief.io)**

---

## How It Works

```
1. Sign up via Google or GitHub OAuth
2. Get a managed inbox (you@mail.mailbrief.io)
3. Forward your newsletters to that address
4. AI extracts key points, topics, action items, and sentiment from each email
5. Open your dashboard to read a pre-computed daily digest
```

No IMAP, no passwords, no browser extensions. Just email forwarding.

## Features

**AI-Powered Intelligence**
- Structured extraction from every email: summaries, key points, topics, links, action items, sentiment
- Daily and weekly digests with diff highlighting (new/dropped senders and topics)
- Smart triage: AI decides which emails are worth extracting, skips transactional noise
- Multi-provider LLM with automatic Anthropic/OpenAI failover
- BYOK (bring your own API key) support

**Inbox Experience**
- Split-pane inline preview on desktop (list left, preview right)
- Email threading with conversation view
- Full-text search with saved filter presets
- Command palette (`Cmd+K`) for navigation, quick filters, and actions
- Keyboard shortcuts (`j`/`k` navigate, `x` select, `e` archive, `s` star, `/` search, `?` help)
- Real-time updates via Server-Sent Events (no polling delay)

**Sender Intelligence**
- Auto-categorization with manual overrides
- Frequency tracking and last-seen timestamps
- Unsubscribe link detection and subscription management
- Feedback-based filtering (thumbs up/down influences future digests)

**Dashboard Analytics**
- Weekly email volume chart
- Topic cloud and category distribution
- Email arrival time heatmap (hour x day-of-week)
- Reading time estimates

**Organisation**
- User-defined labels and folders
- Quiet hours (defer processing during a defined window)
- Auto-archive rules (by category, sender, or age)
- Bulk actions (batch archive, delete, mark read/unread)
- Data export (JSON/CSV) and digest PDF export

**Delivery**
- Digest email delivery via Resend
- Webhook/Slack integration (POST digest summaries to any URL)
- Weekly activity summary email (email counts, top senders, trends)
- Public digest sharing via tokenised links
- Browser notifications for new arrivals
- Favicon unread badge

**Security**
- Email encryption at rest (AES-256-GCM)
- Session token rotation (hourly)
- Per-user rate limiting
- Content Security Policy headers
- Secrets stored in Google Secret Manager

## Architecture

```
                         +-----------------+
                         |   Cloudflare    |
  Inbound email -------> | Email Worker    |
                         +--------+--------+
                                  |
                                  | POST /webhook/email
                                  v
+-----------------+      +--------+--------+      +------------------+
|   Next.js 16    | <--> |   Go Backend    | <--> |   PostgreSQL     |
|   (Frontend)    |      |   (API + LLM)   |      |   (Neon)         |
+-----------------+      +-----------------+      +------------------+
  Better Auth               Anthropic / OpenAI
  React 19                  enmime (email parser)
  Tailwind CSS 4            Resend (outbound email)
  shadcn/ui                 pgx/v5 (database)
```

**Data flow:**
1. Cloudflare Email Worker receives inbound email and POSTs raw RFC 5322 to the Go backend
2. Go backend parses the email, enqueues it for background processing
3. Background worker triages the email (cheap LLM call), then extracts structured data (expensive LLM call)
4. SSE broker pushes real-time notifications to connected browsers
5. Digest scheduler (Cloud Scheduler cron) generates daily/weekly summaries via LLM
6. Frontend authenticates via Better Auth, proxies API requests to Go backend with session tokens

## Tech Stack

| Layer | Technology |
|-------|-----------|
| **Frontend** | Next.js 16, React 19, Tailwind CSS 4, shadcn/ui, Radix UI |
| **Auth** | Better Auth (Google + GitHub OAuth) with hourly token rotation |
| **Backend** | Go 1.24, stdlib `net/http` router, `log/slog` |
| **Database** | PostgreSQL (Neon), `pgx/v5` with tuned connection pool |
| **Email Parsing** | `enmime` (RFC 5322) |
| **AI** | Anthropic Claude + OpenAI (pluggable provider with automatic fallback) |
| **Real-time** | Server-Sent Events (Go broker + frontend EventSource) |
| **Email Sending** | Resend |
| **Email Routing** | Cloudflare Email Workers |
| **Testing** | Vitest + Testing Library (unit), Playwright (E2E), CI via GitHub Actions |
| **API Docs** | OpenAPI 3.1 spec served at `/docs` via Scalar UI |

## Project Structure

```
taran/
├── backend/               # Go API server, email processing, LLM integration
├── frontend/              # Next.js dashboard, auth, UI
├── .github/workflows/     # CI: backend tests, frontend lint/test, E2E with PostgreSQL
├── Makefile               # Dev lifecycle commands
├── INSTALL.md             # Full deployment guide (step-by-step, beginner-friendly)
└── ROADMAP.md             # Feature roadmap
```

See [backend/README.md](backend/README.md) and [frontend/README.md](frontend/README.md) for detailed documentation of each service.

## Quick Start (Local Development)

**Prerequisites:** Go 1.24+, Node.js 22+, Docker

```bash
# 1. Start PostgreSQL
make db

# 2. Configure environment
cp backend/.env.example backend/.env    # Fill in API keys and DB connection
cp frontend/.env.example frontend/.env  # Fill in auth provider credentials

# 3. Start everything
make start        # Builds and starts backend (:8080) + frontend (:3002)

# 4. Seed test data (optional, requires a signed-in user)
make seed         # Creates 30 emails, 2 digests, 3 labels

# Other commands
make stop         # Stop both services
make restart      # Stop then start
```

Frontend at `http://localhost:3002`, backend API at `http://localhost:8080`.

## Deployment

| Component | Platform |
|-----------|----------|
| **Frontend** | [Vercel](https://vercel.com) (auto-deploys on push to `main`) |
| **Backend** | Google Cloud Run (auto-deploys via Cloud Build trigger) |
| **Database** | [Neon](https://neon.tech) (serverless PostgreSQL) |
| **Email Routing** | [Cloudflare](https://cloudflare.com) Email Workers |
| **Digest Scheduler** | Google Cloud Scheduler (hourly cron) |
| **Outbound Email** | [Resend](https://resend.com) |
| **Secrets** | Google Secret Manager |

For a complete step-by-step deployment guide, see **[INSTALL.md](INSTALL.md)**.

## License

[MIT License](LICENSE)
