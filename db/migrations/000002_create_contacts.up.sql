CREATE TABLE contacts (
    id uuid PRIMARY KEY,
    external_contact_id text NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT contacts_external_contact_id_key UNIQUE (external_contact_id)
);
