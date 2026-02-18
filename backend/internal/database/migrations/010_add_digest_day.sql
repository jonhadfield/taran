-- Add configurable day-of-week for weekly digests (0=Sunday..6=Saturday, default 1=Monday)
ALTER TABLE user_preference
    ADD COLUMN IF NOT EXISTS digest_day INTEGER NOT NULL DEFAULT 1;

ALTER TABLE user_preference DROP CONSTRAINT IF EXISTS user_preference_day_check;
ALTER TABLE user_preference ADD CONSTRAINT user_preference_day_check
    CHECK(digest_day >= 0 AND digest_day <= 6);
