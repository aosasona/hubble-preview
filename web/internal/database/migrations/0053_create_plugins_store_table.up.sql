-- Make plugin_identifier UNIQUE
ALTER TABLE installed_plugins
ADD CONSTRAINT unique_plugin_identifier UNIQUE (plugin_identifier);

-- create the plugins_kv table
CREATE TABLE IF NOT EXISTS plugins_kv (
	id uuid NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
	plugin_id text NOT NULL references installed_plugins(plugin_identifier) ON DELETE CASCADE,
	key text NOT NULL,
	value jsonb NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ DEFAULT NULL,
	UNIQUE (plugin_id, key)
);

-- Trigger to set the updated_at column
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON plugins_kv
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

