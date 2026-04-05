-- Store List-Unsubscribe header data parsed from inbound emails
ALTER TABLE email ADD COLUMN IF NOT EXISTS unsubscribe_url TEXT NOT NULL DEFAULT '';
ALTER TABLE email ADD COLUMN IF NOT EXISTS unsubscribe_mailto TEXT NOT NULL DEFAULT '';
