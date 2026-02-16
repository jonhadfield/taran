CREATE TABLE IF NOT EXISTS invite (
    id              TEXT PRIMARY KEY,
    email           TEXT NOT NULL UNIQUE,
    invited_by      TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    accepted_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_invite_email ON invite(email);
