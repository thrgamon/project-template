-- name: CreateUser :one
INSERT INTO users (email)
VALUES ($1)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: DeleteUserByID :exec
DELETE FROM users
WHERE id = $1;
