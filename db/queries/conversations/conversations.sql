-- name: FindConversationWithoutMessages :one
SELECT id, status, user_id, contact_id, created_at, updated_at, finished_at
FROM conversations
WHERE id = $1;

-- name: FindOpenConversationWithMessageExternalIDsByContactID :one
SELECT id, status, user_id, contact_id, created_at, updated_at, finished_at
FROM conversations
WHERE contact_id = $1
  AND status IN ('pending', 'assigned')
LIMIT 1;

-- name: FindConversationWithContactUnreadMessagesByID :one
SELECT id, status, user_id, contact_id, created_at, updated_at, finished_at
FROM conversations
WHERE id = $1;

-- name: UpsertConversation :exec
INSERT INTO conversations (id, status, user_id, contact_id, created_at, updated_at, finished_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status,
    user_id = EXCLUDED.user_id,
    contact_id = EXCLUDED.contact_id,
    updated_at = EXCLUDED.updated_at,
    finished_at = EXCLUDED.finished_at;
