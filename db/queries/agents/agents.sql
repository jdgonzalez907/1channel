-- name: FindAgentByID :one
SELECT id, created_at
FROM agents
WHERE id = $1;
