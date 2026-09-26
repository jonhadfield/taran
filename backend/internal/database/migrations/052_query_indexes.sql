-- Indexes for queries that had none, or whose index had the wrong leading
-- column. All four were found by reading the queries and confirmed against
-- production index-usage statistics, which have never been reset and so cover
-- the database's whole lifetime.
--
-- CREATE INDEX CONCURRENTLY is deliberately not used: migrations run inside a
-- transaction and CONCURRENTLY is not permitted there. These tables are small
-- enough that the brief lock is not a concern.

-- email.from_address is never queried on its own. Every use pairs it with
-- user_id (GetSenderDetail, CountBySenderWeek, the List filter, ListSenders
-- and ListSubscriptions), so a single-column index on from_address can never
-- be the best choice. Production confirms it: zero scans, ever.
--
-- This is the same correction migration 045 already made for message_id.
CREATE INDEX IF NOT EXISTS idx_email_user_from_address ON email(user_id, from_address);
DROP INDEX IF EXISTS idx_email_from_address;

-- digest_item.email_id and .extraction_id are both foreign keys with
-- ON DELETE CASCADE, and Postgres does not index the referencing side
-- automatically. Without these, deleting an email or an extraction
-- sequentially scans digest_item to find the rows to cascade to.
CREATE INDEX IF NOT EXISTS idx_digest_item_email_id ON digest_item(email_id);
CREATE INDEX IF NOT EXISTS idx_digest_item_extraction_id ON digest_item(extraction_id);

-- The only token_usage index is (user_id, created_at), which cannot serve a
-- range scan that does not filter by user. Three queries do exactly that: the
-- daily global total and two admin trends. token_usage gains a row per LLM
-- call, so it grows faster than any other table here.
CREATE INDEX IF NOT EXISTS idx_token_usage_created_at ON token_usage(created_at);

-- The scheduler looks for unsent digests on every tick:
--   WHERE sent_at IS NULL AND generated_at < $1 ORDER BY generated_at
-- Neither column was indexed. A partial index suits this well because rows
-- leave it once they are sent, so it stays close to empty however many
-- digests accumulate.
CREATE INDEX IF NOT EXISTS idx_digest_unsent ON digest(generated_at) WHERE sent_at IS NULL;
