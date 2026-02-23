CREATE TABLE IF NOT EXISTS digest_feedback (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL,
    digest_id  TEXT NOT NULL REFERENCES digest(id) ON DELETE CASCADE,
    rating     TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT digest_feedback_rating_check CHECK(rating IN ('useful', 'not_useful')),
    UNIQUE(user_id, digest_id)
);
CREATE INDEX IF NOT EXISTS idx_digest_feedback_digest ON digest_feedback(digest_id);
