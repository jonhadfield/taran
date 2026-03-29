-- Add encrypted flag to email table
ALTER TABLE email ADD COLUMN IF NOT EXISTS encrypted BOOLEAN NOT NULL DEFAULT FALSE;

-- Update FTS trigger to stop indexing text_body (will be encrypted for new emails)
CREATE OR REPLACE FUNCTION email_search_tsv_trigger() RETURNS trigger AS $$
BEGIN
    NEW.search_tsv :=
        setweight(to_tsvector('english', COALESCE(NEW.subject, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.from_name, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.from_address, '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Recreate trigger without text_body in the column list
DROP TRIGGER IF EXISTS trg_email_search_tsv ON email;
CREATE TRIGGER trg_email_search_tsv
    BEFORE INSERT OR UPDATE OF subject, from_name, from_address
    ON email
    FOR EACH ROW
    EXECUTE FUNCTION email_search_tsv_trigger();
