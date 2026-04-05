-- User preferences (one row per user)
CREATE TABLE IF NOT EXISTS user_preference (
    user_id             TEXT PRIMARY KEY,
    digest_email        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Fix: allow 'skipped' status from triage feature
ALTER TABLE email DROP CONSTRAINT IF EXISTS email_status_check;
ALTER TABLE email ADD CONSTRAINT email_status_check
    CHECK(status IN ('pending','processing','processed','failed','skipped'));
