-- RFC 8058: only senders advertising List-Unsubscribe-Post may be sent an
-- automated one-click POST. Without this we POST to any URL a sender supplies.
ALTER TABLE email ADD COLUMN IF NOT EXISTS unsubscribe_post BOOLEAN NOT NULL DEFAULT FALSE;
