DROP INDEX IF EXISTS idx_instance_admins_user_active;
ALTER TABLE instance_admins DROP COLUMN IF EXISTS deleted_at;
-- instance_id is left nullable; re-adding NOT NULL could fail if NULLs exist.
