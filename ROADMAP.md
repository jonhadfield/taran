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

## Not Started

### Features

#### Custom Digest Templates
- **Effort**: High
- **Impact**: Medium
- **Description**: Move beyond detailed/concise to user-defined digest formats. Let users choose what sections appear (e.g., only action items and links, or summaries only), reorder sections, and set content density per section.
- **Scope**: Backend template schema and LLM prompt customization, frontend template builder UI

#### Slack/Webhook Integration
- **Effort**: Medium
- **Impact**: Medium
- **Description**: Push digest summaries to Slack channels or arbitrary webhook URLs. Users configure destination in settings.
- **Scope**: Backend webhook dispatcher, Slack formatting, frontend config UI

#### Email Rules Engine
- **Effort**: High
- **Impact**: Medium
- **Description**: "If sender matches X, auto-label Y and skip digest" — more powerful than current auto-archive. Condition-action pairs with AND/OR logic.
- **Scope**: Backend rules table, evaluation in worker pipeline, frontend rule builder

#### Digest PDF Export
- **Effort**: Low
- **Impact**: Low
- **Description**: Download digests as formatted PDFs for offline reading or sharing.
- **Scope**: Backend PDF generation endpoint, frontend download button

#### Multi-Inbox
- **Effort**: High
- **Impact**: Medium
- **Description**: Let users create multiple inboxes (e.g., work vs personal newsletters) with separate digest schedules and preferences.
- **Scope**: Backend inbox scoping, frontend inbox switcher, per-inbox settings

### Backend / Infrastructure

#### Background Encryption Migration
- **Effort**: Low
- **Impact**: Medium
- **Description**: Encrypt existing plaintext email bodies. Currently only new emails are encrypted at rest; a background job would migrate historical data.
- **Scope**: Backend migration worker, progress tracking, graceful batching


#### Cloudflare Worker Retry Queue
- **Effort**: Medium
- **Impact**: Medium
- **Description**: If Cloud Run is temporarily down, the Cloudflare email worker currently rejects the email. A retry queue in the worker (using Durable Objects or Queues) would buffer and retry.
- **Scope**: Cloudflare Worker rewrite, queue configuration

#### Database Connection Pool Tuning
- **Effort**: Low
- **Impact**: Low
- **Description**: Configure pgx pool size based on Cloud Run instance count and concurrency settings.
- **Scope**: Backend config, pool configuration

#### Audit Log
- **Effort**: Medium
- **Impact**: Low
- **Description**: Track admin actions (invite, retry, batch operations) for accountability and debugging.
- **Scope**: Backend audit table, middleware, admin UI log viewer

### Security

#### Content Security Policy Headers
- **Effort**: Low
- **Impact**: High
- **Description**: Add CSP headers to prevent XSS beyond the current HTML sanitisation. Restrict script sources, style sources, and frame ancestors.
- **Scope**: Backend security middleware

#### Session Token Rotation
- **Effort**: Medium
- **Impact**: Medium
- **Description**: Rotate session tokens periodically (e.g., every 24 hours) to limit the window of a stolen token. On rotation, the old token is invalidated and a new one is issued transparently.
- **Scope**: Backend session middleware, Better Auth configuration, cookie refresh logic

#### Per-Key API Rate Limiting
- **Effort**: Low
- **Impact**: Medium
- **Description**: Separate rate limits for webhook traffic vs user API traffic. Webhook bursts (email flood) shouldn't block user API requests.
- **Scope**: Backend middleware, separate token buckets by auth type

### Developer Experience

#### OpenAPI Spec
- **Effort**: Medium
- **Impact**: Medium
- **Description**: Hand-written OpenAPI 3.1 spec (the standard formerly known as Swagger) served at `/docs` via Scalar UI. Swagger is the old name — OpenAPI is the spec, Swagger/Scalar/Redoc are UIs that render it. For a plain `net/http` backend, a hand-written YAML spec is cleaner than annotation-based generation (swaggo/swag).
- **Scope**: Backend YAML spec, `/docs` HTML endpoint, Dockerfile update

## Future Considerations

These are not prioritised but worth tracking for later evaluation:

- **Mobile app** — Native iOS/Android or PWA
- **Collaborative accounts** — Shared inboxes for teams
- **Two-factor authentication** — Beyond OAuth2 provider 2FA
- **Digest scheduling on specific dates** — Custom calendar-based scheduling
- **Drag-and-drop label management** — Reorder labels, drag emails into label groups
- **Dashboard widget customisation** — Let users show/hide or reorder dashboard cards
