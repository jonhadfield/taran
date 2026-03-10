-- Add category column to sender_preference for user-override of auto-detected category
ALTER TABLE sender_preference ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT '';

-- Add excluded_categories to user_preference for configurable digest category filtering
ALTER TABLE user_preference ADD COLUMN IF NOT EXISTS excluded_categories JSONB NOT NULL DEFAULT '["notification","transactional","marketing"]';
