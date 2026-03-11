-- Track retry attempts for failed email extractions
ALTER TABLE email ADD COLUMN IF NOT EXISTS retry_count INTEGER NOT NULL DEFAULT 0;
