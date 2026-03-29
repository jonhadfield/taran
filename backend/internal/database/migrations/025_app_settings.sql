CREATE TABLE IF NOT EXISTS app_setting (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO app_setting (key, value) VALUES ('default_monthly_token_limit', '500000')
ON CONFLICT (key) DO NOTHING;
