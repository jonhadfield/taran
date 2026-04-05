-- Composite index for the inbox list query: WHERE user_id = ? ORDER BY received_at DESC
-- Replaces the need for a separate sort step on every page load.
CREATE INDEX IF NOT EXISTS idx_email_user_received
    ON email (user_id, received_at DESC);

-- The single-column user_id index is now redundant (the composite covers it).
DROP INDEX IF EXISTS idx_email_user_id;
