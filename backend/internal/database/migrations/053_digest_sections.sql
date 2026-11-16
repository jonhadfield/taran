-- Per-user digest section toggles (custom digest template, first slice).
-- Defaults keep today's behaviour: highlights, topics, and included emails on.
ALTER TABLE user_preference
    ADD COLUMN IF NOT EXISTS digest_include_highlights BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS digest_include_topics BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS digest_include_emails BOOLEAN NOT NULL DEFAULT TRUE;
