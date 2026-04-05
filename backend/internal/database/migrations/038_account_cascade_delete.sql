-- Replace email.email_account_id FK with ON DELETE CASCADE
-- so deleting an email_account hard-deletes all associated emails
-- (which then cascades to extractions, feedback, attachments, labels, digest_items)

ALTER TABLE email DROP CONSTRAINT IF EXISTS email_email_account_id_fkey;
ALTER TABLE email ADD CONSTRAINT email_email_account_id_fkey
    FOREIGN KEY (email_account_id) REFERENCES email_account(id) ON DELETE CASCADE;
