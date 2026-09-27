-- name: FindUserByID :one
SELECT id, created_at
FROM users
WHERE id = $1;

-- name: CreateUser :exec
INSERT INTO users (id, created_at)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE
SET created_at = EXCLUDED.created_at;
