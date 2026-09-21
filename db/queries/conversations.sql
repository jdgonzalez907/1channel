-- name: FindConversationByID :many
SELECT
    sqlc.embed(c),
    sqlc.embed(m)
FROM conversations c
JOIN messages m ON m.conversation_id = c.id
WHERE c.id = sqlc.arg('id')
ORDER BY m.registered_at;

-- name: FindLastOpenConversationByContactID :many
SELECT
    sqlc.embed(c),
    sqlc.embed(m)
FROM conversations c
JOIN messages m ON m.conversation_id = c.id
WHERE c.contact_id = sqlc.arg('contact_id')
  AND (c.status NOT IN ('expired', 'resolved') OR c.finished_at > now())
ORDER BY c.created_at DESC, m.registered_at;

-- name: FindWithSpecificMessageByExternalID :one
SELECT
    sqlc.embed(c),
    sqlc.embed(m)
FROM messages m
JOIN conversations c ON c.id = m.conversation_id
WHERE m.external_id = sqlc.arg('external_id');

-- name: FindWithSpecificMessageByMessageID :one
SELECT
    sqlc.embed(c),
    sqlc.embed(m)
FROM messages m
JOIN conversations c ON c.id = m.conversation_id
WHERE m.id = sqlc.arg('message_id');

-- name: UpsertConversation :exec
INSERT INTO conversations (id, status, unread_count, agent_id, contact_id, created_at, updated_at, finished_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status,
    unread_count = EXCLUDED.unread_count,
    agent_id = EXCLUDED.agent_id,
    updated_at = EXCLUDED.updated_at,
    finished_at = EXCLUDED.finished_at;

-- name: BatchUpsertMessages :exec
INSERT INTO messages (id, conversation_id, external_id, text, message_type, status, agent_id, contact_id, registered_at, updated_at, sent_at, delivered_at, read_at, failed_at, deleted_at)
SELECT
    unnest(sqlc.arg('ids')::uuid[]),
    unnest(sqlc.arg('conversation_ids')::uuid[]),
    unnest(sqlc.arg('external_ids')::text[]),
    unnest(sqlc.arg('texts')::text[]),
    unnest(sqlc.arg('message_types')::text[]),
    unnest(sqlc.arg('statuses')::text[]),
    unnest(sqlc.arg('agent_ids')::uuid[]),
    unnest(sqlc.arg('contact_ids')::uuid[]),
    unnest(sqlc.arg('registered_ats')::timestamptz[]),
    unnest(sqlc.arg('updated_ats')::timestamptz[]),
    unnest(sqlc.arg('sent_ats')::timestamptz[]),
    unnest(sqlc.arg('delivered_ats')::timestamptz[]),
    unnest(sqlc.arg('read_ats')::timestamptz[]),
    unnest(sqlc.arg('failed_ats')::timestamptz[]),
    unnest(sqlc.arg('deleted_ats')::timestamptz[])
ON CONFLICT (id) DO UPDATE SET
    external_id = EXCLUDED.external_id,
    text = EXCLUDED.text,
    status = EXCLUDED.status,
    updated_at = EXCLUDED.updated_at,
    sent_at = EXCLUDED.sent_at,
    delivered_at = EXCLUDED.delivered_at,
    read_at = EXCLUDED.read_at,
    failed_at = EXCLUDED.failed_at,
    deleted_at = EXCLUDED.deleted_at;
