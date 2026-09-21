ALTER TABLE messages RENAME COLUMN created_at TO registered_at;
ALTER TABLE messages ADD COLUMN sent_at TIMESTAMPTZ;
ALTER TABLE messages ADD COLUMN delivered_at TIMESTAMPTZ;
ALTER TABLE messages ADD COLUMN failed_at TIMESTAMPTZ;
