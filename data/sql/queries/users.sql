-- name: GetUserByEmail :one
SELECT
    *
FROM
    users u
WHERE
    u.email = $1
LIMIT
    1;
