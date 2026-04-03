-- Add index on digest.share_token for shared digest lookups
CREATE INDEX IF NOT EXISTS idx_digest_share_token ON digest(share_token) WHERE share_token IS NOT NULL;

-- Add index on email.from_address for sender queries
CREATE INDEX IF NOT EXISTS idx_email_from_address ON email(from_address);
