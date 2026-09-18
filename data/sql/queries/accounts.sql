-- name: GetAccountByID :one
SELECT
    *
FROM
    accounts a
WHERE
    a.id = $1
LIMIT
    1;
