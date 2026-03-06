CREATE TABLE IF NOT EXISTS email_attachment (
    id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL REFERENCES email(id) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    content_type TEXT NOT NULL DEFAULT 'application/octet-stream',
    size_bytes INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_email_attachment_email_id ON email_attachment(email_id);
