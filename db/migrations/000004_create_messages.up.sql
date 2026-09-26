CREATE TABLE messages (
    id uuid PRIMARY KEY,
    conversation_id uuid NOT NULL,
    status text NOT NULL,
    type text NOT NULL,
    text text,
    agent_id uuid,
    contact_id uuid,
    external_id text,
    sent_at timestamptz NOT NULL,
    read_at timestamptz,
    edited_at timestamptz,
    deleted_at timestamptz,
    CONSTRAINT messages_conversation_id_fkey FOREIGN KEY (conversation_id)
        REFERENCES conversations (id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT messages_agent_id_fkey FOREIGN KEY (agent_id)
        REFERENCES agents (id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT messages_contact_id_fkey FOREIGN KEY (contact_id)
        REFERENCES contacts (id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT messages_status_check CHECK (
        status IN ('sent', 'read', 'deleted', 'failed')
    ),
    CONSTRAINT messages_type_check CHECK (type IN ('text')),
    CONSTRAINT messages_owner_check CHECK (
        (agent_id IS NULL) <> (contact_id IS NULL)
    )
);

CREATE INDEX messages_conversation_id_idx ON messages (conversation_id);

CREATE UNIQUE INDEX messages_external_id_key
    ON messages (external_id)
    WHERE external_id IS NOT NULL;
