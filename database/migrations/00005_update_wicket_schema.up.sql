-- 1. Drop the redundant is_runout column
ALTER TABLE balls DROP COLUMN IF EXISTS is_runout;

-- 2. Rename caught_by to fielder_id to make it universal for all fielding actions
ALTER TABLE balls RENAME COLUMN caught_by TO fielder_id;

-- (Optional but recommended) Rename the foreign key constraint if you want it to look clean in the DB.
-- If you don't know the exact constraint name, Postgres usually names it "balls_caught_by_fkey".
ALTER TABLE balls RENAME CONSTRAINT balls_caught_by_fkey TO balls_fielder_id_fkey;

-- Remove redundant delivery_type column
ALTER TABLE balls DROP COLUMN IF EXISTS delivery_type;