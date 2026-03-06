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
- **Status**: Not started
- **Effort**: Medium
- **Impact**: High
- **Description**: Token usage is tracked but not enforced. Add per-user usage limits, a usage dashboard in settings, and admin-level cost monitoring. Critical for sustainable scaling.
- **Scope**: Backend usage tracking table, rate limiting by token budget, admin cost dashboard, user usage display in settings

### 2.2 Usage Analytics Dashboard
- **Status**: Not started
- **Effort**: Medium
- **Impact**: Medium
- **Description**: Beyond basic admin stats, track digest open rates, feedback trends over time, per-user engagement metrics, and email processing success rates. Helps understand product-market fit and identify issues.
- **Scope**: Backend analytics aggregation endpoints, admin analytics UI with charts/trends

### 2.3 Improved Search (Full-Text)
- **Status**: Not started
- **Effort**: Medium
- **Impact**: Medium
- **Description**: Replace basic substring matching with PostgreSQL full-text search using `tsvector`/`tsquery`. Support quoted phrases, boolean operators, and search across email subject, body, and extraction summaries.
- **Scope**: Database migration for tsvector columns/indexes, updated search query logic, frontend search syntax hints

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
- **Status**: Not started
- **Effort**: Low
- **Impact**: Medium
- **Description**: The forwarding guide is hardcoded for a few email providers. Expand to dynamic, per-provider instructions (Gmail, Outlook, Yahoo, Apple Mail, ProtonMail, etc.) with screenshots or step-by-step walkthroughs. Reduces onboarding friction significantly.
- **Scope**: Frontend guide content expansion, possibly a guided wizard component

## Future Considerations

These are not prioritized but worth tracking for later evaluation:

- **Email threading / conversation view** — Group related emails into threads
- **Mobile app** — Native iOS/Android or PWA
- **API webhooks for third-party integrations** — Let users push digest data to Slack, Notion, etc.
- **Collaborative accounts** — Shared inboxes for teams
- **Two-factor authentication** — Beyond OAuth2 provider 2FA
- **Digest scheduling on specific dates** — Custom calendar-based scheduling
- **Email encryption at rest** — Enhanced data protection
