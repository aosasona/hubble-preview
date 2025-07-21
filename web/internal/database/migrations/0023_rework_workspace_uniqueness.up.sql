-- We don't need these columns anymore
ALTER TABLE workspaces
DROP COLUMN namespaced_name,
DROP COLUMN slug,
ADD CONSTRAINT uq_user_workspaces UNIQUE (owner_id, display_name);

