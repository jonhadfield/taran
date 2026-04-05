-- Weekly activity summary emails (lighter than digests)
CREATE TABLE IF NOT EXISTS weekly_summary (
    id            TEXT PRIMARY KEY,
    user_id       TEXT NOT NULL,
    period_start  TIMESTAMPTZ NOT NULL,
    period_end    TIMESTAMPTZ NOT NULL,
    email_count   INTEGER NOT NULL DEFAULT 0,
    top_senders   JSONB NOT NULL DEFAULT '[]',
    categories    JSONB NOT NULL DEFAULT '{}',
    action_items  INTEGER NOT NULL DEFAULT 0,
    sent_at       TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, period_start)
);
CREATE INDEX IF NOT EXISTS idx_weekly_summary_user_id ON weekly_summary(user_id);

-- Add weekly summary opt-in to user preferences (defaults to true)
ALTER TABLE user_preference ADD COLUMN IF NOT EXISTS weekly_summary BOOLEAN NOT NULL DEFAULT true;
