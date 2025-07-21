ALTER TABLE collections
DROP COLUMN slug;

ALTER TABLE collections
ADD COLUMN slug TEXT GENERATED ALWAYS AS (
	slugify(name) || '-' || workspace_id
) STORED;

