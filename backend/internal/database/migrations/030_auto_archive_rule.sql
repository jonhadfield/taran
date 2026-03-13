CREATE TABLE IF NOT EXISTS auto_archive_rule (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    rule_type TEXT NOT NULL,
    rule_value TEXT NOT NULL,
    archive_after_days INTEGER NOT NULL DEFAULT 7,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, rule_type, rule_value)
);
CREATE INDEX IF NOT EXISTS idx_auto_archive_rule_user ON auto_archive_rule(user_id);
