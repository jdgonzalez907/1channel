-- name: FindMessageByExternalID :one
SELECT id, conversation_id, status, type, text, agent_id, contact_id, external_id, sent_at, read_at, edited_at, deleted_at
FROM messages
WHERE external_id = $1;

-- name: ListMessagesWithExternalIDByConversation :many
SELECT id, conversation_id, status, type, text, agent_id, contact_id, external_id, sent_at, read_at, edited_at, deleted_at
FROM messages
WHERE conversation_id = $1
  AND external_id IS NOT NULL
ORDER BY sent_at, id;

-- name: ListUnreadContactMessagesByConversation :many
SELECT id, conversation_id, status, type, text, agent_id, contact_id, external_id, sent_at, read_at, edited_at, deleted_at
FROM messages
WHERE conversation_id = $1
  AND contact_id IS NOT NULL
  AND read_at IS NULL
ORDER BY sent_at, id;

-- name: UpsertMessage :exec
INSERT INTO messages (id, conversation_id, status, type, text, agent_id, contact_id, external_id, sent_at, read_at, edited_at, deleted_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status,
    type = EXCLUDED.type,
    text = EXCLUDED.text,
    agent_id = EXCLUDED.agent_id,
    contact_id = EXCLUDED.contact_id,
    external_id = EXCLUDED.external_id,
    sent_at = EXCLUDED.sent_at,
    read_at = EXCLUDED.read_at,
    edited_at = EXCLUDED.edited_at,
    deleted_at = EXCLUDED.deleted_at;
