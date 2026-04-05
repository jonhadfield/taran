-- Webhook delivery preferences for digest notifications
ALTER TABLE user_preference ADD COLUMN IF NOT EXISTS digest_webhook BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE user_preference ADD COLUMN IF NOT EXISTS webhook_url TEXT NOT NULL DEFAULT '';
