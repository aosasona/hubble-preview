ALTER TABLE workspaces
DROP COLUMN namespaced_name;

ALTER TABLE workspaces
ADD COLUMN namespaced_name VARCHAR(255)
GENERATED ALWAYS AS (owner_id || '.' || slug) STORED;

ALTER TABLE workspaces
ADD CONSTRAINT uq_workspaces_namespaced_name UNIQUE (namespaced_name);

