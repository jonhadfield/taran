ALTER TABLE user_preference ADD COLUMN IF NOT EXISTS digest_style TEXT NOT NULL DEFAULT 'detailed';
