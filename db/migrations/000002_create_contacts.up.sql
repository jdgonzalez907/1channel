CREATE TABLE contacts (
    id                  UUID         PRIMARY KEY,
    external_contact_id VARCHAR(255) NOT NULL,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_contacts_external_id ON contacts (external_contact_id);
