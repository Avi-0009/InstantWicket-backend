-- Remove the strict unique constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_phone_no_key;

-- Add the partial unique index that ignores archived users
CREATE UNIQUE INDEX IF NOT EXISTS idx_active_users_phone ON users (phone_no) WHERE archived_at IS NULL;