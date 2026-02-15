CREATE TABLE IF NOT EXISTS email_feedback (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    email_id TEXT NOT NULL REFERENCES email(id) ON DELETE CASCADE,
    rating TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT feedback_rating_check CHECK(rating IN ('useful', 'not_useful')),
    UNIQUE(user_id, email_id)
);
CREATE INDEX IF NOT EXISTS idx_email_feedback_email ON email_feedback(email_id);
