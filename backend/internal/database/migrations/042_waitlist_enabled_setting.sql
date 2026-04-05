-- Add waitlist_enabled app setting (default: disabled)
INSERT INTO app_setting (key, value, updated_at)
VALUES ('waitlist_enabled', 'false', NOW())
ON CONFLICT (key) DO NOTHING;
