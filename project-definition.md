# Taran — Email Digest Dashboard

## Overview

Taran is a multi-user web application that provides each user with a managed email inbox. Users subscribe to newsletters using their assigned address, and the platform uses AI to extract structured data and generate daily/weekly digest summaries.

## Goals

- **Newsletter-first**: Optimized for reading and summarizing newsletters, but handles all email types
- **AI-powered extraction**: Use LLMs to produce structured data from raw email content
- **Digest-oriented**: Pre-computed daily and weekly digests, ready when you open the dashboard
- **Multi-user**: Users sign up via OAuth, each gets their own managed inbox
- **Self-hosted**: Runs on your own infrastructure via Docker

## Architecture

```
User signs up → Better Auth (Next.js) → user record in shared PostgreSQL DB
                                       → managed inbox assigned (user@mail.taran.app)

Email arrives → Cloudflare Email Worker → POST /webhook/email (Go backend)
              → Parse RFC 5322 → LLM extraction → store in DB

User opens app → Next.js → GET /api/* (Go backend, session auth)
              → Go validates Better Auth session → returns data

Cron trigger → digest generator → LLM summarization → store digest
```

## Application Stack

### Backend (Go)
- **Webhook endpoint**: Receives raw email from Cloudflare Email Workers via HTTP POST
- **Email processor**: Parses RFC 5322 email content (HTML/plain text), cleans it, and sends it to the LLM layer for analysis
- **LLM layer**: Pluggable provider interface supporting multiple backends (Anthropic Claude, OpenAI, Ollama/local models). Swappable via configuration
- **Background worker**: Channel-based async processing of incoming emails with bounded concurrency
- **Digest generator**: Pre-computes daily/weekly digests on a configurable cron schedule
- **REST API**: Serves processed data and digests to the frontend
- **Session validation**: Validates Better Auth sessions from the shared database

### Frontend (Next.js)
- **Framework**: Next.js with React
- **Styling**: Tailwind CSS for responsive, visually appealing design
- **Authentication**: Better Auth with Google/GitHub/SSO OAuth providers
- **Primary view**: Daily/weekly digest — an aggregated briefing of newsletter content and other emails
- **Additional views**: Per-email detail view, account management, settings

### Email Ingestion (Cloudflare)
- **Cloudflare Email Routing**: Catch-all rule routes all inbound email to an Email Worker
- **Email Worker**: Reads raw email stream, POSTs to Go backend webhook with shared secret auth
- **Address resolution**: Go backend maps the "to" address to a user account

### Database (PostgreSQL)
- **Single shared database** between Better Auth (Next.js) and Go backend
- **Better Auth manages**: `user`, `session`, `account` tables
- **Go backend manages**: `email_account`, `email`, `extraction`, `digest`, `digest_item` tables
- **JSONB columns** for structured array data (key points, topics, links, etc.)

## Data Extraction

The LLM extracts the following structured data from each email:

| Field | Description |
|-------|-------------|
| **Summary** | Brief plain-language summary of the email content |
| **Key points** | Bullet-point list of main takeaways |
| **Topics** | Tags/categories for grouping and filtering |
| **Links** | Important URLs extracted from the content |
| **Action items** | Any calls-to-action, deadlines, or tasks mentioned |
| **Sentiment** | Overall tone (informational, urgent, promotional, etc.) |
| **Source category** | Classification of the sender (newsletter, personal, transactional, etc.) |

## User Model

- **Multi-user**: Users sign up via Better Auth (Google/GitHub/SSO)
- **Managed inbox**: Each user is assigned an email address (e.g. `username@mail.taran.app`)
- **Unified digest**: All emails to a user's inbox feed into their digest

## Digest Features

- **Email delivery**: Optionally send the pre-computed digest to a configured email address
- **Item interactions**: Users can mark digest items as read, star/favourite important ones, and archive old items
- **Responsive design**: Mobile-friendly responsive web design (not a PWA)

## Project Structure

```
taran/
├── backend/           # Go: webhook, email processing, LLM integration, API server
│   ├── cmd/taran/     # Application entry point
│   ├── internal/      # All application packages
│   └── migrations/    # PostgreSQL migration SQL files
├── frontend/          # Next.js: dashboard, auth, digest views, settings UI
├── worker/            # Cloudflare Email Worker (TypeScript)
├── docker-compose.yaml
├── Makefile
└── README.md
```

## Configuration

Key configurable values (via environment variables):
- PostgreSQL connection string
- Webhook shared secret (for Cloudflare Worker → Go backend auth)
- LLM provider selection and API keys
- Digest schedule (cron expression)
- Email domain for managed inboxes
- Better Auth OAuth provider credentials (Google, GitHub)
