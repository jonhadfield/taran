ALTER TABLE user_preference ADD COLUMN IF NOT EXISTS interest_keywords JSONB NOT NULL DEFAULT '[]';
ALTER TABLE user_preference ADD COLUMN IF NOT EXISTS exclusion_keywords JSONB NOT NULL DEFAULT '[]';
