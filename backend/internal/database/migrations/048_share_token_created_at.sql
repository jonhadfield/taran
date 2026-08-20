-- Share links expire 30 days after sharing, not 30 days after the digest was
-- generated. Backfill existing shared rows from created_at so their expiry is
-- unchanged by this migration.
ALTER TABLE digest ADD COLUMN IF NOT EXISTS share_token_created_at TIMESTAMPTZ;
UPDATE digest SET share_token_created_at = created_at
 WHERE share_token IS NOT NULL AND share_token_created_at IS NULL;
