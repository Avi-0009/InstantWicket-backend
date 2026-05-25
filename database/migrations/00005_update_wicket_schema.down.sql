-- 1. Revert fielder_id back to caught_by
ALTER TABLE balls RENAME COLUMN fielder_id TO caught_by;

-- Revert the constraint name
ALTER TABLE balls RENAME CONSTRAINT balls_fielder_id_fkey TO balls_caught_by_fkey;

-- 2. Add the is_runout column back
ALTER TABLE balls ADD COLUMN is_runout BOOLEAN DEFAULT FALSE;

-- Bring back delivery_type if rolling back
ALTER TABLE balls ADD COLUMN delivery_type TEXT DEFAULT 'normal';