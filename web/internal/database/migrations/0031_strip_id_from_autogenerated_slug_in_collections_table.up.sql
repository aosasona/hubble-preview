ALTER TABLE collections
DROP COLUMN slug;

-- We don't need the uniqueness factor since:
-- a. collections are scoped to a workspace
-- b. the slug is generated from the name
ALTER TABLE collections
ADD COLUMN slug TEXT GENERATED ALWAYS AS (
	slugify(name)
) STORED;

