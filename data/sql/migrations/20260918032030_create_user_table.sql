-- +goose Up
CREATE TABLE users (
    id             BIGINT      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    uid            TEXT,
    name           TEXT,
    email          TEXT        UNIQUE NOT NULL,
    email_verified BOOLEAN     DEFAULT FALSE NOT NULL,
    image          TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX users_uid_idx ON users (uid);

-- +goose Down
DROP TABLE IF EXISTS users;
