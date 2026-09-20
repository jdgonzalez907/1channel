-- name: FindAgentByID :one
SELECT sqlc.embed(a)
FROM agents a
WHERE a.id = $1;
