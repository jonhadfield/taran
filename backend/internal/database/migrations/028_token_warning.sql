ALTER TABLE user_preference ADD COLUMN IF NOT EXISTS token_warning_sent_at TIMESTAMPTZ;
