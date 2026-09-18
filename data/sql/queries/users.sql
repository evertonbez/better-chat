-- name: GetUserByEmail :one
SELECT
    *
FROM
    users u
WHERE
    u.email = $1
LIMIT
    1;

-- name: GetUserByID :one
SELECT
    *
FROM
    users u
WHERE
    u.id = $1
LIMIT
    1;

-- name: GetUserByUID :one
SELECT
    *
FROM
    users u
WHERE
    u.uid = $1
LIMIT
    1;

-- name: CreateUser :one
INSERT INTO
    users (uid, name, email, email_verified, image)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: UpdateUser :one
UPDATE users
SET
    name = COALESCE(sqlc.narg('name'), name),
    image = COALESCE(sqlc.narg('image'), image),
    updated_at = NOW()
WHERE
    id = sqlc.arg('id')
RETURNING
    *;
