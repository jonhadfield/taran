-- Raw webhook payloads contain the full RFC 5322 message, i.e. the same body
-- text that is encrypted in email.text_body/html_body. Track which rows are
-- encrypted so existing plaintext rows stay readable during rollout.
ALTER TABLE webhook_payload ADD COLUMN IF NOT EXISTS encrypted BOOLEAN NOT NULL DEFAULT FALSE;
