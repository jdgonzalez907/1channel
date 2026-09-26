CREATE TABLE conversations (
    id uuid PRIMARY KEY,
    status text NOT NULL,
    agent_id uuid,
    contact_id uuid,
    created_at timestamptz NOT NULL,
    updated_at timestamptz,
    finished_at timestamptz,
    CONSTRAINT conversations_agent_id_fkey FOREIGN KEY (agent_id)
        REFERENCES agents (id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT conversations_contact_id_fkey FOREIGN KEY (contact_id)
        REFERENCES contacts (id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT conversations_status_check CHECK (
        status IN ('pending', 'assigned', 'expired', 'resolved')
    ),
    CONSTRAINT conversations_owner_check CHECK (
        agent_id IS NOT NULL OR contact_id IS NOT NULL
    ),
    CONSTRAINT conversations_finished_at_check CHECK (
        status IN ('pending', 'assigned') OR finished_at IS NOT NULL
    )
);

CREATE UNIQUE INDEX conversations_one_open_per_contact
    ON conversations (contact_id)
    WHERE status IN ('pending', 'assigned')
      AND contact_id IS NOT NULL;
