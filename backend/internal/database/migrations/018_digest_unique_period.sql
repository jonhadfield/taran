ALTER TABLE digest ADD CONSTRAINT uq_digest_user_period
    UNIQUE (user_id, period_start, period_end);
