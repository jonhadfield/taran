-- Inbound webhook dedupe and threading lookups are scoped by user_id, so the
-- single-column message_id index no longer matches the query shape.
CREATE INDEX IF NOT EXISTS idx_email_user_message_id ON email(user_id, message_id);
DROP INDEX IF EXISTS idx_email_message_id;
