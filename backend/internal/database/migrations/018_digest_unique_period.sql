-- Remove duplicate digests (keep the earliest per user/period), then add unique constraint.
DELETE FROM digest d
USING digest d2
WHERE d.user_id = d2.user_id
  AND d.period_start = d2.period_start
  AND d.period_end = d2.period_end
  AND d.created_at > d2.created_at;

ALTER TABLE digest ADD CONSTRAINT uq_digest_user_period
    UNIQUE (user_id, period_start, period_end);
