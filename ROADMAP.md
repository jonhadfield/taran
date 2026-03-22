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

## Not Started

### UI/UX

#### Onboarding Checklist
- **Effort**: Low
- **Impact**: Medium
- **Description**: Persistent checklist after signup (forward first email, configure digest schedule, etc.) that dismisses when all steps are complete.
- **Scope**: Frontend component, backend preference flag

#### Digest Diff Highlighting
- **Effort**: Low
- **Impact**: Low
- **Description**: Visually highlight new content that appeared in a digest compared to the previous one (beyond the current new/dropped senders and topics).
- **Scope**: Frontend diff rendering on digest detail page

#### Empty State Illustrations
- **Effort**: Low
- **Impact**: Low
- **Description**: Replace generic icons with more engaging illustrations or animated SVGs for zero-state inbox, digests, and senders pages.
- **Scope**: Frontend design assets

### Features

#### Custom Digest Templates
- **Effort**: High
- **Impact**: Medium
- **Description**: Move beyond detailed/concise to user-defined digest formats. Let users choose what sections appear (e.g., only action items and links, or summaries only), reorder sections, and set content density per section.
- **Scope**: Backend template schema and LLM prompt customization, frontend template builder UI

#### Weekly Activity Summary Email
- **Effort**: Medium
- **Impact**: Medium
- **Description**: Lighter than digest — email counts, top senders, action items pending, trends vs previous week. Sent weekly regardless of digest schedule.
- **Scope**: Backend summary generator, Resend HTML template, user preference toggle

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

#### Per-User Rate Limiting
- **Effort**: Medium
- **Impact**: Medium
- **Description**: Current rate limiter is global. Per-user limits would prevent one heavy user from starving others.
- **Scope**: Backend middleware, token bucket per user ID

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

#### E2E Tests with Playwright
- **Effort**: High
- **Impact**: High
- **Description**: Test the full login → forward email → view inbox → generate digest flow in a real browser. Run in CI against a test database.
- **Scope**: Playwright test suite, CI configuration, test fixtures

#### Seed Script
- **Effort**: Low
- **Impact**: Medium
- **Description**: Populate local dev with realistic test data — users, emails, extractions, digests, labels — so developers can work on the UI without manual setup.
- **Scope**: Backend CLI command or script

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
