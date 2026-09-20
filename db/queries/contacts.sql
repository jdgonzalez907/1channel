-- name: FindContactByExternalID :one
SELECT sqlc.embed(c)
FROM contacts c
WHERE c.external_contact_id = $1;

-- name: UpsertContact :exec
INSERT INTO contacts (id, external_contact_id, created_at)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE SET
    external_contact_id = EXCLUDED.external_contact_id;
