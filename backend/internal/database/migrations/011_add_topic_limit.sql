-- Add configurable topic cloud limit to user preferences
ALTER TABLE user_preference ADD COLUMN IF NOT EXISTS topic_limit INTEGER NOT NULL DEFAULT 15;
