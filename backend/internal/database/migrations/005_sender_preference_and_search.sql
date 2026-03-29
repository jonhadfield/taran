CREATE TABLE IF NOT EXISTS sender_preference (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    from_address TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'normal',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT sender_pref_status_check CHECK(status IN ('normal', 'muted', 'blocked', 'favorite')),
    UNIQUE(user_id, from_address)
);

CREATE INDEX IF NOT EXISTS idx_sender_pref_user ON sender_preference(user_id);
