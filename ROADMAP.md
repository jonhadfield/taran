# MailBrief Roadmap

## Completed

### High Impact / User Experience
- **Bulk Email Actions** — Batch archive, delete, mark read/unread from multi-select inbox UI
- **Digest Preview** — Draft/preview step before finalizing digest generation
- **Attachment Display** — Store and display attachments in email detail view
- **Data Export** — Export emails and digests as JSON/CSV for GDPR compliance
- **Browser Notifications** — Notify on new email arrivals when tab is unfocused
- **Breadcrumb Navigation** — Inbox / Subject breadcrumbs on detail pages
- **Settings Navigation** — Sticky pill bar with scroll tracking, grouped sections

### Operational / Business
- **LLM Cost Controls** — Per-user monthly/daily token limits, 80% warning emails, usage dashboard, admin cost monitoring
- **Usage Analytics Dashboard** — Admin pipeline health, feedback summary, 8-week trend charts
- **Improved Search (Full-Text)** — PostgreSQL FTS on subject/sender/extraction fields, sort by relevance
- **Multi-Provider LLM Fallback** — Automatic Anthropic/OpenAI failover with FallbackProvider
- **Webhook Reliability** — Dead letter queue, admin replay/retry, pipeline health monitoring
- **Email Encryption at Rest** — AES-256-GCM encryption of email bodies in the repository layer
- **Performance Indexes** — Composite indexes on hot query paths, N+1 fix in thread handler
- **Dashboard Query Optimisation** — Lightweight count queries, CTE-based sender category lookups
- **Server-Side Sanitisation** — Moved sanitize-html from client bundle to server component

### Growth / Retention
- **Email Labels / Folders** — User-defined labels with inbox filtering
- **Onboarding Email Provider Guides** — 9 provider forwarding guides with pro tips
- **Mobile Responsive Polish** — Touch-friendly inbox, responsive layouts across all pages
- **Token Limit Warning Email** — Automated 80% usage notification via Resend
- **Keyboard Shortcuts** — j/k navigate, x select, e archive, s star, / search, Enter open, Esc clear
- **Quiet Hours** — Defer processing during user-defined window
- **Auto-Archive Rules** — Archive by category/sender after N days
- **Category Distribution Chart** — Animated horizontal bars by category on dashboard
- **Email Arrival Time Heatmap** — Hour x day-of-week grid with hover tooltips
- **Reading Time Estimates** — Per-email estimate at 200 WPM
- **Saved Searches / Filter Presets** — Save/load/delete filter presets (max 20 per user)
- **Unsubscribe Tracking** — Surface senders with unsubscribe links, manage subscriptions view
- **Sender Detail Page** — Last seen, avg frequency, inbox link
- **Digest Comparison** — New/dropped senders and topics vs previous digest
- **Email Threading** — Conversation view grouping related emails
- **Animated Dashboard Charts** — Staggered bar grow animations with hover tooltips
- **Inline Email Preview** — Split-pane view on desktop: email list left, preview right, keyboard-driven
- **Command Palette** — Cmd+K overlay with navigation, quick filters, and actions (theme toggle, export, generate digest)
- **Server-Sent Events** — Real-time push replacing 60s polling; SSE broker in Go backend, auto-reconnecting EventSource on frontend
- **Onboarding Checklist** — Persistent dashboard checklist (create inbox, receive email, generate digest, configure settings) with progress bar, dismissible
- **Digest Diff Highlighting** — Inline visual diffs: new topics get sparkle icon + green ring, new-sender emails get green left border + badge
- **Empty State Illustrations** — Custom SVG illustrations for inbox, digests, senders, dashboard, search, and subscriptions empty states
- **Seed Script** — `make seed` populates local dev with 30 emails, 2 digests, 3 labels for realistic testing
- **E2E Tests with Playwright** — Workflow tests for inbox, digests, command palette, onboarding; CI job with PostgreSQL + backend
- **Per-User Rate Limiting** — Per-user token bucket (5 req/s, 20 burst) after auth, in addition to existing per-IP limiter
- **Weekly Activity Summary Email** — Sunday email with email count, top senders, category breakdown, action items; opt-in via preferences
- **Session Token Rotation** — Tokens rotate every hour; backend generates new token, proxy sets new cookie with HMAC signature
- **Content Security Policy Headers** — CSP headers on all responses, special handling for /docs CDN scripts
- **Split Rate Limiting** — Separate rate limits for API (10 req/s) vs webhook/cron (50 req/s) traffic
- **OpenAPI Spec** — Full OpenAPI 3.1 spec covering all 59 endpoints, served at /docs via Scalar UI
- **Slack/Webhook Integration** — POST digest summaries to user-configured webhook URLs; settings UI with URL input
- **Digest PDF Export** — Download any digest as a formatted PDF (title, summary, highlights, topics, included emails)
- **DB Connection Pool Tuning** — Configured pgx pool: 10 max / 2 min conns, 30m lifetime, 5m idle, 30s health check
- **Audit Log** — Middleware auto-logs admin write operations; table + admin UI with scrollable log viewer

- **Keyboard Shortcut Help Overlay** — `?` key shows all shortcuts in a modal (inbox navigation, actions, command palette)
- **Favicon Unread Badge** — Unread email count rendered as a red badge on the browser tab favicon
- **Weekly Summary Settings Toggle** — UI toggle in delivery settings for the weekly activity summary preference
- **Webhook Test Button** — "Send test" button in settings with backend endpoint to verify webhook URL
- **Test Results Gitignore** — Added `test-results/` and `playwright-report/` to `.gitignore`
- **Preference Handler WeeklySummary** — Added WeeklySummary to the PATCH /api/preferences update request

- **SSE Reconnection Indicator** — Toast notification on SSE reconnection with status tracking in useEventSource hook
- **Error Boundary** — Dashboard error.tsx with "Try Again" and "Dashboard" recovery buttons
- **Orphaned Digest Cleanup** — Sweeper deletes digests with zero remaining items every 5 minutes

## Not Started

### Features
- **Custom Digest Templates** — User-defined digest formats: choose sections, reorder, set density
- **Email Rules Engine** — Condition-action pairs ("if sender matches X, auto-label Y and skip digest")
- **Multi-Inbox** — Multiple inboxes per user with separate digest schedules and preferences

### Infrastructure
- **Background Encryption Migration** — Encrypt existing plaintext email bodies (only new emails are encrypted)
- **Cloudflare Worker Retry Queue** — Buffer emails when Cloud Run is down using Durable Objects or Queues

## Future Considerations

These are not prioritised but worth tracking for later evaluation:

- **Mobile app** — Native iOS/Android or PWA
- **Collaborative accounts** — Shared inboxes for teams
- **Two-factor authentication** — Beyond OAuth2 provider 2FA
- **Digest scheduling on specific dates** — Custom calendar-based scheduling
- **Drag-and-drop label management** — Reorder labels, drag emails into label groups
- **Dashboard widget customisation** — Let users show/hide or reorder dashboard cards
