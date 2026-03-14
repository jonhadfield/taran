# MailBrief Roadmap

## Priority 1: High Impact / User Experience

### 1.1 Bulk Email Actions
- **Status**: Done
- **Effort**: Low
- **Impact**: High
- **Description**: Add batch operations to the inbox — select multiple emails, then archive, delete, or mark read/unread in one action. Essential for users managing high-volume newsletters.
- **Scope**: Frontend multi-select UI, backend batch endpoint (`PATCH /api/emails/batch`)

### 1.2 Digest Preview
- **Status**: Done
- **Effort**: Medium
- **Impact**: High
- **Description**: Allow users to preview a digest before it's finalized and sent. Introduces a draft/preview step so users can see what the digest will contain, edit if needed, and then confirm generation. Reduces wasted LLM calls and gives users more control.
- **Scope**: Backend draft generation endpoint, frontend preview UI, confirm/discard flow

### 1.3 Attachment Display
- **Status**: Done
- **Effort**: Medium
- **Impact**: Medium
- **Description**: Attachments are parsed during email ingestion but not stored or displayed. Store attachments (with size limits) and show them in the email detail view so users don't lose context.
- **Scope**: Backend attachment storage (object storage or DB), email detail UI update, download endpoint

### 1.4 Data Export
- **Status**: Done
- **Effort**: Medium
- **Impact**: Medium
- **Description**: Allow users to export their emails and digests (JSON, CSV, or both). Important for user trust and GDPR compliance. Could also support digest export as PDF.
- **Scope**: Backend export endpoint (`GET /api/export`), frontend settings integration

## Priority 2: Operational / Business

### 2.1 LLM Cost Controls
- **Status**: Done
- **Effort**: Medium
- **Impact**: High
- **Description**: Per-user monthly and daily token limits with enforcement (skip/defer on exceed), 80% warning emails, usage dashboard with 30-day history chart, projected usage, operation breakdown, admin cost monitoring with per-user limit editing.
- **Scope**: Token usage tracking, monthly + daily rate limiting, admin cost dashboard, user usage display with history

### 2.2 Usage Analytics Dashboard
- **Status**: Done
- **Effort**: Medium
- **Impact**: Medium
- **Description**: Admin dashboard with processing pipeline health (processed/failed/skipped/pending counts with success rate), user feedback summary (useful/not-useful ratio), 8-week trend charts for emails, digests, and token usage.
- **Scope**: Backend analytics queries in admin stats endpoint, frontend mini bar charts, processing pipeline card, feedback card

### 2.3 Improved Search (Full-Text)
- **Status**: Done
- **Effort**: Medium
- **Impact**: Medium
- **Description**: PostgreSQL full-text search with `tsvector`/`tsquery` on email subject, body, sender name/address, and AI-extracted summaries/key points/action items. Sort options (newest/oldest/relevance), search result count display, and search-aware empty states.
- **Scope**: Extended search query to include extraction content, from_name ILIKE fallback, sort parameter, frontend sort dropdown and UX improvements

## Priority 3: Growth / Retention

### 3.1 Custom Digest Templates
- **Status**: Not started
- **Effort**: High
- **Impact**: Medium
- **Description**: Move beyond detailed/concise to user-defined digest formats. Let users choose what sections appear (e.g., only action items and links, or summaries only), reorder sections, and set content density per section.
- **Scope**: Backend template schema and LLM prompt customization, frontend template builder UI

### 3.2 Email Labels / Folders
- **Status**: Not started
- **Effort**: Medium
- **Impact**: Medium
- **Description**: The inbox is flat (star/archive only). Add user-defined labels or folders so users can organize emails by project, priority, or custom categories. Labels could also feed into digest filtering.
- **Scope**: Backend labels table, label CRUD endpoints, inbox filter by label, frontend label management UI

### 3.3 Onboarding Email Provider Guides
- **Status**: Done
- **Effort**: Low
- **Impact**: Medium
- **Description**: Expanded forwarding guide from 4 to 9 providers: Gmail, Outlook, Yahoo Mail, Apple Mail (iCloud), ProtonMail, Fastmail, Zoho Mail, Hey.com, and direct subscription. Each provider includes pro tips for selective forwarding.
- **Scope**: Frontend forwarding guide component expansion with per-provider tips

### 3.4 Mobile Responsive Polish
- **Status**: Done
- **Effort**: Low
- **Impact**: Medium
- **Description**: Mobile-first fixes across all pages: inbox row actions always visible on touch devices, bulk action bar icons-only on mobile, keyboard shortcuts hidden on mobile, HTML email content overflow prevention, senders page two-row layout on mobile with visible dropdowns and category badges, email detail responsive header and subject sizing, better touch targets.
- **Scope**: Frontend CSS/layout changes across inbox, senders, email detail, and bulk action components

### 3.5 Token Limit Warning Email
- **Status**: Done
- **Effort**: Low
- **Impact**: Medium
- **Description**: Automated email notification at 80% monthly token usage with HTML progress bar, formatted usage stats, and link to settings. Sends once per month via Resend, triggered after each email extraction in the worker pipeline.
- **Scope**: Already implemented in mailer (SendTokenWarning), worker (checkTokenWarning), and database (SetTokenWarningSent)

## Future Considerations

These are not prioritized but worth tracking for later evaluation:

- **Email threading / conversation view** — Group related emails into threads
- **Mobile app** — Native iOS/Android or PWA
- **API webhooks for third-party integrations** — Let users push digest data to Slack, Notion, etc.
- **Collaborative accounts** — Shared inboxes for teams
- **Two-factor authentication** — Beyond OAuth2 provider 2FA
- **Digest scheduling on specific dates** — Custom calendar-based scheduling
- **Email encryption at rest** — Enhanced data protection
