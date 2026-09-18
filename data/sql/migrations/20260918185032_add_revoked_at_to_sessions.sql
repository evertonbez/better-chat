-- +goose Up
ALTER TABLE sessions
ADD COLUMN revoked_at TIMESTAMPTZ;

CREATE INDEX sessions_live_token_idx ON sessions (token)
WHERE
    revoked_at IS NULL;

CREATE INDEX sessions_live_userId_idx ON sessions (user_id)
WHERE
    revoked_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS sessions_live_userId_idx;

DROP INDEX IF EXISTS sessions_live_token_idx;

ALTER TABLE sessions
DROP COLUMN revoked_at;
