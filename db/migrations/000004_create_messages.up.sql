CREATE TABLE messages (
    id uuid PRIMARY KEY,
    conversation_id uuid NOT NULL,
    status text NOT NULL,
    type text NOT NULL,
    text text,
    user_id uuid,
    contact_id uuid,
    external_id text,
    sent_at timestamptz NOT NULL,
    read_at timestamptz,
    edited_at timestamptz,
    deleted_at timestamptz,
    CONSTRAINT messages_conversation_id_fkey FOREIGN KEY (conversation_id)
        REFERENCES conversations (id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT messages_user_id_fkey FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT messages_contact_id_fkey FOREIGN KEY (contact_id)
        REFERENCES contacts (id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT messages_status_check CHECK (
        status IN ('sent', 'read', 'deleted', 'failed')
    ),
    CONSTRAINT messages_type_check CHECK (type IN ('text')),
    CONSTRAINT messages_owner_check CHECK (
        (user_id IS NULL) <> (contact_id IS NULL)
    )
);

CREATE INDEX messages_conversation_id_idx
    ON messages (conversation_id, sent_at DESC, id DESC);

CREATE UNIQUE INDEX messages_external_id_key
    ON messages (external_id)
    WHERE external_id IS NOT NULL;
