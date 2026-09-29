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

-- name: RefreshConversationLastMessage :exec
UPDATE conversations c
SET last_message_id = last_message.id,
    last_message_at = last_message.sent_at,
    unread_count = unread_count.cnt
FROM (
    SELECT m.id, m.sent_at
    FROM messages m
    WHERE m.conversation_id = $1
    ORDER BY m.sent_at DESC, m.id DESC
    LIMIT 1
) AS last_message,
(
    SELECT COUNT(*) AS cnt
    FROM messages m
    WHERE m.conversation_id = $1
      AND m.contact_id IS NOT NULL
      AND m.read_at IS NULL
) AS unread_count
WHERE c.id = $1;

-- name: FindConversationWithContactByID :one
SELECT
    c.id,
    c.status,
    c.user_id,
    c.created_at,
    c.updated_at,
    c.finished_at,
    c.unread_count,
    ct.id AS contact_id,
    ct.external_contact_id,
    ct.display_name,
    pi.id AS pi_id,
    COALESCE(pi.identification_number, '') AS identification_number,
    pi.first_name,
    pi.last_name,
    pi.phone_number,
    pi.email,
    pi.address,
    pi.created_at AS pi_created_at,
    pi.updated_at AS pi_updated_at
FROM conversations c
JOIN contacts ct ON ct.id = c.contact_id
LEFT JOIN personal_information pi ON pi.id = ct.personal_information_id
WHERE c.id = $1;

-- name: ListConversationsForAgent :many
SELECT
    c.id,
    c.status,
    c.last_message_at,
    c.unread_count,
    ct.id AS contact_id,
    ct.external_contact_id,
    ct.display_name,
    pi.first_name,
    pi.last_name,
    lm.status AS last_message_status,
    lm.text AS last_message_text,
    CASE WHEN lm.user_id IS NOT NULL THEN 'agent' ELSE 'contact' END AS last_message_owner
FROM conversations c
JOIN contacts ct ON ct.id = c.contact_id
LEFT JOIN personal_information pi ON pi.id = ct.personal_information_id
JOIN messages lm ON lm.id = c.last_message_id
WHERE (c.user_id IS NULL OR c.user_id = sqlc.arg('agent_id'))
  AND c.status = ANY(sqlc.arg('statuses')::text[])
  AND (sqlc.narg('contact_id')::uuid IS NULL OR c.contact_id = sqlc.narg('contact_id'))
  AND (
      sqlc.narg('before_sent_at')::timestamptz IS NULL
      OR (c.last_message_at, c.id) < (sqlc.narg('before_sent_at')::timestamptz, sqlc.narg('before_id')::uuid)
  )
ORDER BY c.last_message_at DESC, c.id DESC
LIMIT sqlc.arg('page_size')::int + 1;
