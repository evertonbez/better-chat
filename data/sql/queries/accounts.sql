-- name: GetAccountByID :one
SELECT
    *
FROM
    accounts a
WHERE
    a.id = $1
LIMIT
    1;

-- name: GetAccountByProvider :one
SELECT
    *
FROM
    accounts a
WHERE
    a.user_id = $1
    AND a.provider_id = $2
LIMIT
    1;

-- name: CreateAccount :one
INSERT INTO
    accounts (
        provider_id,
        user_id,
        password,
        access_token,
        refresh_token,
        id_token,
        access_token_expires_at,
        refresh_token_expires_at
    )
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING
    *;

-- name: UpdateAccountPassword :exec
UPDATE accounts
SET
    password = $3,
    updated_at = NOW()
WHERE
    user_id = $1
    AND provider_id = $2;
