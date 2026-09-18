-- +goose Up
CREATE TABLE sessions (
    id         BIGINT      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    token      TEXT        UNIQUE NOT NULL,
    user_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX sessions_userId_idx ON sessions USING BTREE (user_id);

CREATE INDEX sessions_expiresAt_idx ON sessions USING BTREE (expires_at);

-- +goose Down
DROP TABLE sessions;
