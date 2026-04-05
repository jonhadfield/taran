ALTER TABLE sender_preference ADD COLUMN IF NOT EXISTS unsubscribed_at TIMESTAMPTZ;
