CREATE TABLE IF NOT EXISTS label (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    color      TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, name)
);
CREATE INDEX IF NOT EXISTS idx_label_user ON label(user_id);

CREATE TABLE IF NOT EXISTS email_label (
    email_id TEXT NOT NULL REFERENCES email(id) ON DELETE CASCADE,
    label_id TEXT NOT NULL REFERENCES label(id) ON DELETE CASCADE,
    PRIMARY KEY (email_id, label_id)
);
CREATE INDEX IF NOT EXISTS idx_email_label_label ON email_label(label_id);
