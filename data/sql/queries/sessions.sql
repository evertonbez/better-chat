-- name: GetSessionByID :one
SELECT
    *
FROM
    sessions s
WHERE
    s.id = $1
LIMIT
    1;

-- name: GetLiveSessionWithUserByToken :one
SELECT
    sqlc.embed(s),
    sqlc.embed(u)
FROM
    sessions s
    JOIN users u ON u.id = s.user_id
WHERE
    s.token = $1
    AND s.revoked_at IS NULL
LIMIT
    1;

-- name: CreateSession :one
INSERT INTO
    sessions (token, user_id, expires_at, ip_address, user_agent)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: RefreshSession :one
UPDATE sessions
SET
    expires_at = $2,
    updated_at = NOW()
WHERE
    token = $1
    AND revoked_at IS NULL
RETURNING
    *;

-- name: RevokeSessionByToken :execrows
UPDATE sessions
SET
    revoked_at = NOW(),
    updated_at = NOW()
WHERE
    token = $1
    AND revoked_at IS NULL;

-- name: RevokeSessionsByUserID :execrows
UPDATE sessions
SET
    revoked_at = NOW(),
    updated_at = NOW()
WHERE
    user_id = $1
    AND revoked_at IS NULL;

-- name: ListLiveSessionTokensByUserID :many
SELECT
    s.token
FROM
    sessions s
WHERE
    s.user_id = $1
    AND s.revoked_at IS NULL
    AND s.expires_at > NOW();

-- name: ListSessionsByUserID :many
SELECT
    *
FROM
    sessions s
WHERE
    s.user_id = $1
ORDER BY
    s.created_at DESC;
