CREATE TABLE conversations (
    id uuid PRIMARY KEY,
    status text NOT NULL,
    user_id uuid,
    contact_id uuid,
    created_at timestamptz NOT NULL,
    updated_at timestamptz,
    finished_at timestamptz,
    last_message_at timestamptz,
    last_message_id uuid,
    unread_count integer NOT NULL DEFAULT 0,
    CONSTRAINT conversations_user_id_fkey FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT conversations_contact_id_fkey FOREIGN KEY (contact_id)
        REFERENCES contacts (id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT conversations_status_check CHECK (
        status IN ('pending', 'assigned', 'expired', 'resolved')
    ),
    CONSTRAINT conversations_owner_check CHECK (
        user_id IS NOT NULL OR contact_id IS NOT NULL
    ),
    CONSTRAINT conversations_finished_at_check CHECK (
        status IN ('pending', 'assigned') OR finished_at IS NOT NULL
    ),
    CONSTRAINT conversations_unread_count_check CHECK (unread_count >= 0)
);

CREATE UNIQUE INDEX conversations_one_open_per_contact
    ON conversations (contact_id)
    WHERE status IN ('pending', 'assigned')
      AND contact_id IS NOT NULL;

CREATE INDEX conversations_user_last_message_idx
    ON conversations (user_id, last_message_at DESC, id DESC);

CREATE INDEX conversations_status_last_message_idx
    ON conversations (status, last_message_at DESC, id DESC);

CREATE INDEX conversations_contact_id_idx
    ON conversations (contact_id);
