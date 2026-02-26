ALTER TABLE user_preference ADD COLUMN IF NOT EXISTS color_theme TEXT NOT NULL DEFAULT 'neutral';
