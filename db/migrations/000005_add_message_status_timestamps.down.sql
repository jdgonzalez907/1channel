ALTER TABLE messages DROP COLUMN IF EXISTS failed_at;
ALTER TABLE messages DROP COLUMN IF EXISTS delivered_at;
ALTER TABLE messages DROP COLUMN IF EXISTS sent_at;
ALTER TABLE messages RENAME COLUMN registered_at TO created_at;
