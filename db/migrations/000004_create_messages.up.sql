CREATE TABLE messages (
    id              UUID         PRIMARY KEY,
    conversation_id UUID         NOT NULL REFERENCES conversations(id) ON DELETE RESTRICT,
    external_id     VARCHAR(255),
    text            TEXT         NOT NULL DEFAULT '',
    message_type    VARCHAR(20)  NOT NULL DEFAULT 'text'
                    CHECK (message_type IN ('text')),
    status          VARCHAR(20)  NOT NULL DEFAULT 'registered'
                    CHECK (status IN ('registered', 'sent', 'delivered', 'read', 'deleted')),
    agent_id        UUID         REFERENCES agents(id) ON DELETE RESTRICT,
    contact_id      UUID         REFERENCES contacts(id) ON DELETE RESTRICT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ,
    read_at         TIMESTAMPTZ
);

CREATE INDEX idx_messages_conversation_created ON messages (conversation_id, created_at);
CREATE UNIQUE INDEX idx_messages_external_id   ON messages (external_id);
