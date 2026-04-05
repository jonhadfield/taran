-- Add tsvector column for full-text search on email
ALTER TABLE email ADD COLUMN IF NOT EXISTS search_tsv tsvector;

-- Populate existing rows
UPDATE email SET search_tsv =
    setweight(to_tsvector('english', COALESCE(subject, '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(from_name, '')), 'B') ||
    setweight(to_tsvector('english', COALESCE(from_address, '')), 'B') ||
    setweight(to_tsvector('english', COALESCE(LEFT(text_body, 5000), '')), 'C');

-- GIN index for fast full-text search
CREATE INDEX IF NOT EXISTS idx_email_search_tsv ON email USING GIN (search_tsv);

-- Trigger to keep tsvector up-to-date on insert/update
CREATE OR REPLACE FUNCTION email_search_tsv_trigger() RETURNS trigger AS $$
BEGIN
    NEW.search_tsv :=
        setweight(to_tsvector('english', COALESCE(NEW.subject, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.from_name, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.from_address, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(LEFT(NEW.text_body, 5000), '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_email_search_tsv ON email;
CREATE TRIGGER trg_email_search_tsv
    BEFORE INSERT OR UPDATE OF subject, from_name, from_address, text_body
    ON email
    FOR EACH ROW
    EXECUTE FUNCTION email_search_tsv_trigger();
