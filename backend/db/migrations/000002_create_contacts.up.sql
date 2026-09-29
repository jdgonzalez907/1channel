CREATE TABLE personal_information (
    id uuid PRIMARY KEY,
    identification_number text NOT NULL,
    first_name text,
    last_name text,
    phone_number text,
    email text,
    address text,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT personal_information_identification_number_key UNIQUE (identification_number)
);

CREATE TABLE contacts (
    id uuid PRIMARY KEY,
    external_contact_id text NOT NULL,
    display_name text,
    personal_information_id uuid,
    created_at timestamptz NOT NULL,
    CONSTRAINT contacts_external_contact_id_key UNIQUE (external_contact_id),
    CONSTRAINT contacts_personal_information_id_fkey FOREIGN KEY (personal_information_id)
        REFERENCES personal_information (id) ON DELETE RESTRICT ON UPDATE RESTRICT
);
