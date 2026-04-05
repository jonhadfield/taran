-- Add ON DELETE CASCADE to foreign keys so deleting an email cleans up extractions and digest items.

ALTER TABLE extraction DROP CONSTRAINT IF EXISTS extraction_email_id_fkey;
ALTER TABLE extraction ADD CONSTRAINT extraction_email_id_fkey
    FOREIGN KEY (email_id) REFERENCES email(id) ON DELETE CASCADE;

ALTER TABLE digest_item DROP CONSTRAINT IF EXISTS digest_item_email_id_fkey;
ALTER TABLE digest_item ADD CONSTRAINT digest_item_email_id_fkey
    FOREIGN KEY (email_id) REFERENCES email(id) ON DELETE CASCADE;

ALTER TABLE digest_item DROP CONSTRAINT IF EXISTS digest_item_extraction_id_fkey;
ALTER TABLE digest_item ADD CONSTRAINT digest_item_extraction_id_fkey
    FOREIGN KEY (extraction_id) REFERENCES extraction(id) ON DELETE CASCADE;
