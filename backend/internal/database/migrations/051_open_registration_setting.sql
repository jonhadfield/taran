-- Add open_registration app setting (default: disabled). When enabled, any
-- signed-in user is granted access without an invite.
INSERT INTO app_setting (key, value, updated_at)
VALUES ('open_registration', 'false', NOW())
ON CONFLICT (key) DO NOTHING;
