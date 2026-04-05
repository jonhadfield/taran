-- Token usage ledger: records every LLM call for cost tracking and limits
CREATE TABLE IF NOT EXISTS token_usage (
    id            TEXT PRIMARY KEY,
    user_id       TEXT NOT NULL,
    operation     TEXT NOT NULL,  -- 'triage', 'extract', 'digest'
    provider      TEXT NOT NULL,
    model         TEXT NOT NULL,
    input_tokens  INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens  INTEGER NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_token_usage_user_month
    ON token_usage (user_id, created_at);

-- Track tokens used per digest generation
ALTER TABLE digest ADD COLUMN IF NOT EXISTS tokens_used INTEGER NOT NULL DEFAULT 0;

-- Per-user monthly token limit (0 = unlimited)
ALTER TABLE user_preference ADD COLUMN IF NOT EXISTS monthly_token_limit INTEGER NOT NULL DEFAULT 0;
