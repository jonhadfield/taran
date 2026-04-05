-- Add status_reason to persist why an email was skipped or failed
ALTER TABLE email ADD COLUMN IF NOT EXISTS status_reason TEXT NOT NULL DEFAULT '';
