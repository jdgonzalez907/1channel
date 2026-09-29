-- name: FindContactByID :one
SELECT id, external_contact_id, display_name, personal_information_id, created_at
FROM contacts
WHERE id = $1;

-- name: FindContactByExternalContactID :one
SELECT id, external_contact_id, display_name, personal_information_id, created_at
FROM contacts
WHERE external_contact_id = $1;

-- name: FindContactWithPersonalInformationByID :one
SELECT
    ct.id,
    ct.external_contact_id,
    ct.display_name,
    ct.created_at,
    pi.id AS pi_id,
    COALESCE(pi.identification_number, '') AS identification_number,
    pi.first_name,
    pi.last_name,
    pi.phone_number,
    pi.email,
    pi.address,
    pi.created_at AS pi_created_at,
    pi.updated_at AS pi_updated_at
FROM contacts ct
LEFT JOIN personal_information pi ON pi.id = ct.personal_information_id
WHERE ct.id = $1;

-- name: UpsertContact :exec
INSERT INTO contacts (id, external_contact_id, display_name, personal_information_id, created_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE SET
    external_contact_id = EXCLUDED.external_contact_id,
    display_name = EXCLUDED.display_name,
    personal_information_id = EXCLUDED.personal_information_id;
