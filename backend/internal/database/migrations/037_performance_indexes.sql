-- Composite index for List() filtering by status (used on every inbox page load)
CREATE INDEX IF NOT EXISTS idx_email_user_status_received
    ON email (user_id, status, received_at DESC);

-- Index for retry sweep query (runs every 5 minutes in worker)
CREATE INDEX IF NOT EXISTS idx_email_status_updated
    ON email (status, updated_at) WHERE status = 'failed';

-- GIN index on extraction topics for topic-based inbox filtering
CREATE INDEX IF NOT EXISTS idx_extraction_topics_gin
    ON extraction USING GIN (topics);

-- Index for category filtering via extraction
CREATE INDEX IF NOT EXISTS idx_extraction_source_category
    ON extraction (email_id, source_category);

-- Index for email_label lookups by email (List filter uses email_id first)
CREATE INDEX IF NOT EXISTS idx_email_label_email
    ON email_label (email_id, label_id);

-- Index for email_feedback user queries
CREATE INDEX IF NOT EXISTS idx_email_feedback_user
    ON email_feedback (user_id, email_id);
