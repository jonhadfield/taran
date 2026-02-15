-- Extend user_preference with digest customization
ALTER TABLE user_preference
    ADD COLUMN IF NOT EXISTS digest_frequency TEXT NOT NULL DEFAULT 'daily',
    ADD COLUMN IF NOT EXISTS digest_hour INTEGER NOT NULL DEFAULT 7,
    ADD COLUMN IF NOT EXISTS digest_timezone TEXT NOT NULL DEFAULT 'UTC';

ALTER TABLE user_preference DROP CONSTRAINT IF EXISTS user_preference_frequency_check;
ALTER TABLE user_preference ADD CONSTRAINT user_preference_frequency_check
    CHECK(digest_frequency IN ('daily', 'weekly'));

ALTER TABLE user_preference DROP CONSTRAINT IF EXISTS user_preference_hour_check;
ALTER TABLE user_preference ADD CONSTRAINT user_preference_hour_check
    CHECK(digest_hour >= 0 AND digest_hour <= 23);

-- Share token for public digest links
ALTER TABLE digest ADD COLUMN IF NOT EXISTS share_token TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_digest_share_token ON digest(share_token) WHERE share_token IS NOT NULL;
