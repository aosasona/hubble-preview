ALTER TABLE workspaces
ADD COLUMN slug TEXT GENERATED ALWAYS AS (
	slugify(display_name) || '-' || owner_id
) STORED;

