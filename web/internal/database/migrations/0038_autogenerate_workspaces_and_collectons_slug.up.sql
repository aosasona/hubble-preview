-- Drop existing unique constraints if they exist
ALTER TABLE workspaces DROP CONSTRAINT IF EXISTS workspaces_slug_unique;
ALTER TABLE collections DROP CONSTRAINT IF EXISTS collections_slug_unique;

-- Add a slug column to workspaces and collections
--
-- Drop the previous slug
ALTER TABLE workspaces DROP COLUMN slug;
ALTER TABLE collections DROP COLUMN slug;

-- Add a new slug column
ALTER TABLE workspaces ADD COLUMN slug TEXT DEFAULT NULL;
ALTER TABLE collections ADD COLUMN slug TEXT DEFAULT NULL;

-- Update the slug for all existing workspaces
UPDATE workspaces
SET slug = slugify(display_name)
WHERE slug IS NULL;

-- Update the slug for all existing collections
UPDATE collections
SET slug = slugify(name)
WHERE slug IS NULL;

-- Add a unique constraint to the slug column
ALTER TABLE workspaces ADD CONSTRAINT workspaces_slug_unique UNIQUE (slug);

-- NOTE: collection slugs are not unique across workspaces
ALTER TABLE collections ADD CONSTRAINT collections_slug_unique UNIQUE (slug, workspace_id);

-- Add indexes to the slug columns
CREATE INDEX ON collections (workspace_id);
CREATE INDEX ON collections (slug, workspace_id);

CREATE INDEX ON workspaces (slug);
CREATE INDEX ON workspaces (public_id);

