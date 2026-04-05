CREATE TABLE IF NOT EXISTS user_llm_key (
    id            TEXT PRIMARY KEY,
    user_id       TEXT NOT NULL,
    provider      TEXT NOT NULL,
    encrypted_key BYTEA NOT NULL,
    key_nonce     BYTEA NOT NULL,
    key_hint      TEXT NOT NULL DEFAULT '',
    model         TEXT NOT NULL DEFAULT '',
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_user_llm_key_user_provider
    ON user_llm_key (user_id, provider);
