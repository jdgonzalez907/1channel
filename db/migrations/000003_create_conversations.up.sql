CREATE TABLE conversations (
    id           UUID         PRIMARY KEY,
    status       VARCHAR(20)  NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending', 'assigned', 'expired', 'resolved')),
    unread_count SMALLINT     NOT NULL DEFAULT 0
                 CHECK (unread_count >= 0 AND unread_count <= 100),
    agent_id     UUID         REFERENCES agents(id) ON DELETE RESTRICT,
    contact_id   UUID         NOT NULL REFERENCES contacts(id) ON DELETE RESTRICT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ,
    finished_at  TIMESTAMPTZ
);

CREATE INDEX idx_conversations_contact_status ON conversations (contact_id, status);
CREATE INDEX idx_conversations_agent_status   ON conversations (agent_id, status);
CREATE INDEX idx_conversations_created_at     ON conversations (created_at);
