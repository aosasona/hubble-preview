-- Remove the plugin_privileges table
DROP TABLE IF EXISTS plugin_privileges CASCADE;

-- Embed the privileges into the installed_plugins table
-- I thought about using the metadata column, but we migth need that in the future
ALTER TABLE installed_plugins
ADD COLUMN IF NOT EXISTS privileges jsonb DEFAULT '{}'::jsonb;

