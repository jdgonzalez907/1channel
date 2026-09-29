-- name: FindPersonalInformationByID :one
SELECT id, identification_number, first_name, last_name, phone_number, email, address, created_at, updated_at
FROM personal_information
WHERE id = $1;

-- name: FindPersonalInformationByIdentificationNumber :one
SELECT id, identification_number, first_name, last_name, phone_number, email, address, created_at, updated_at
FROM personal_information
WHERE identification_number = $1;

-- name: UpsertPersonalInformation :one
INSERT INTO personal_information (
    id, identification_number, first_name, last_name, phone_number, email, address, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (identification_number) DO UPDATE SET
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    phone_number = EXCLUDED.phone_number,
    email = EXCLUDED.email,
    address = EXCLUDED.address,
    updated_at = EXCLUDED.updated_at
RETURNING id, identification_number, first_name, last_name, phone_number, email, address, created_at, updated_at;
