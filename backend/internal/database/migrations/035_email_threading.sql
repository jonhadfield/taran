-- Add threading support to emails
ALTER TABLE email ADD COLUMN IF NOT EXISTS in_reply_to TEXT NOT NULL DEFAULT '';
ALTER TABLE email ADD COLUMN IF NOT EXISTS thread_id TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_email_thread_id ON email(thread_id) WHERE thread_id != '';
CREATE INDEX IF NOT EXISTS idx_email_in_reply_to ON email(in_reply_to) WHERE in_reply_to != '';
