-- Email accounts: maps users to their assigned inbound email addresses.
CREATE TABLE IF NOT EXISTS email_account (
    id            TEXT PRIMARY KEY,
    user_id       TEXT NOT NULL,
    email_address TEXT NOT NULL UNIQUE,
    display_name  TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_email_account_user_id ON email_account(user_id);

-- Inbound emails received via webhook.
CREATE TABLE IF NOT EXISTS email (
    id               TEXT PRIMARY KEY,
    user_id          TEXT NOT NULL,
    email_account_id TEXT NOT NULL REFERENCES email_account(id),
    message_id       TEXT NOT NULL DEFAULT '',
    from_address     TEXT NOT NULL,
    from_name        TEXT NOT NULL DEFAULT '',
    to_address       TEXT NOT NULL,
    subject          TEXT NOT NULL DEFAULT '',
    text_body        TEXT NOT NULL DEFAULT '',
    html_body        TEXT NOT NULL DEFAULT '',
    received_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    date_header      TIMESTAMPTZ,
    status           TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','processing','processed','failed')),
    is_read          BOOLEAN NOT NULL DEFAULT FALSE,
    is_starred       BOOLEAN NOT NULL DEFAULT FALSE,
    is_archived      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_email_user_id ON email(user_id);
CREATE INDEX IF NOT EXISTS idx_email_status ON email(status);
CREATE INDEX IF NOT EXISTS idx_email_received_at ON email(received_at);
CREATE INDEX IF NOT EXISTS idx_email_message_id ON email(message_id);
CREATE INDEX IF NOT EXISTS idx_email_account_id ON email(email_account_id);

-- LLM extraction results, one per email.
CREATE TABLE IF NOT EXISTS extraction (
    id              TEXT PRIMARY KEY,
    email_id        TEXT NOT NULL UNIQUE REFERENCES email(id),
    summary         TEXT NOT NULL DEFAULT '',
    key_points      JSONB NOT NULL DEFAULT '[]',
    topics          JSONB NOT NULL DEFAULT '[]',
    links           JSONB NOT NULL DEFAULT '[]',
    action_items    JSONB NOT NULL DEFAULT '[]',
    sentiment       TEXT NOT NULL DEFAULT '',
    source_category TEXT NOT NULL DEFAULT '',
    provider        TEXT NOT NULL,
    model           TEXT NOT NULL,
    tokens_used     INTEGER NOT NULL DEFAULT 0,
    processed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_extraction_email_id ON extraction(email_id);

-- Generated digests.
CREATE TABLE IF NOT EXISTS digest (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL,
    title        TEXT NOT NULL DEFAULT '',
    summary      TEXT NOT NULL DEFAULT '',
    highlights   JSONB NOT NULL DEFAULT '[]',
    top_topics   JSONB NOT NULL DEFAULT '[]',
    period_start TIMESTAMPTZ NOT NULL,
    period_end   TIMESTAMPTZ NOT NULL,
    period_type  TEXT NOT NULL CHECK(period_type IN ('daily','weekly')),
    email_count  INTEGER NOT NULL DEFAULT 0,
    provider     TEXT NOT NULL,
    model        TEXT NOT NULL,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at      TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_digest_user_id ON digest(user_id);
CREATE INDEX IF NOT EXISTS idx_digest_period ON digest(period_start, period_end);
CREATE INDEX IF NOT EXISTS idx_digest_type ON digest(period_type);

-- Links digests to their source emails and extractions.
CREATE TABLE IF NOT EXISTS digest_item (
    id            TEXT PRIMARY KEY,
    digest_id     TEXT NOT NULL REFERENCES digest(id) ON DELETE CASCADE,
    email_id      TEXT NOT NULL REFERENCES email(id),
    extraction_id TEXT NOT NULL REFERENCES extraction(id),
    sort_order    INTEGER NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_digest_item_digest_id ON digest_item(digest_id);
