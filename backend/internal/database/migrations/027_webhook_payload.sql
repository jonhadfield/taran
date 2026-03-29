CREATE TABLE IF NOT EXISTS webhook_payload (
    id          TEXT PRIMARY KEY,
    email_id    TEXT REFERENCES email(id) ON DELETE SET NULL,
    raw_body    BYTEA NOT NULL,
    headers     JSONB NOT NULL DEFAULT '{}',
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    size_bytes  INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_webhook_payload_email_id ON webhook_payload(email_id);
CREATE INDEX IF NOT EXISTS idx_webhook_payload_received_at ON webhook_payload(received_at);
