-- name: GetSessionByID :one
SELECT
    *
FROM
    sessions s
WHERE
    s.id = $1
LIMIT
    1;
