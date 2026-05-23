-- Reverse the process: Drop the partial index
DROP INDEX IF EXISTS idx_active_users_phone;

-- Add back the strict unique constraint
ALTER TABLE users ADD CONSTRAINT users_phone_no_key UNIQUE (phone_no);