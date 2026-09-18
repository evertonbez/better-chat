-- +goose Up
CREATE TABLE accounts (
    id                       BIGINT      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    provider_id              TEXT        NOT NULL,
    user_id                  BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    access_token             TEXT,
    refresh_token            TEXT,
    id_token                 TEXT,
    access_token_expires_at  TIMESTAMPTZ,
    refresh_token_expires_at TIMESTAMPTZ,
    password                 TEXT,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX account_userId_idx ON accounts USING BTREE (user_id);

-- +goose Down
DROP TABLE accounts;
