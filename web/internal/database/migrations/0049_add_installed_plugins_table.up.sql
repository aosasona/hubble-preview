CREATE TYPE plugin_scope AS ENUM ('workspace', 'global');

CREATE TYPE plugin_mode AS ENUM ('on_create', 'background');

CREATE TABLE IF NOT EXISTS installed_plugins (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	plugin_identifier TEXT NOT NULL,
	workspace_id integer REFERENCES workspaces(id) ON DELETE CASCADE,
	source_id UUID REFERENCES plugin_sources(id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	description TEXT,
	scope plugin_scope NOT NULL DEFAULT 'workspace',
	modes plugin_mode[] NOT NULL DEFAULT '{}',
	entry_types TEXT[] NOT NULL,
	version_sha TEXT NOT NULL,

	last_updated_at TIMESTAMPTZ DEFAULT NULL,
	added_at TIMESTAMPTZ DEFAULT NULL,
	updated_at TIMESTAMPTZ DEFAULT NULL,

	metadata JSONB DEFAULT '{}'::jsonb,
	tags TEXT[] DEFAULT '{}',

	UNIQUE (workspace_id, plugin_identifier)
);

COMMENT ON COLUMN installed_plugins.plugin_identifier IS 'A unique identifier for the plugin, this is generated in the system as a hash from the source data and the workspace itself. It is also used to identify local files related to the plugin.';

COMMENT ON COLUMN installed_plugins.version_sha IS 'The SHA-256 of the WASM file provided as part of the build';

-- Indexes for perf
CREATE INDEX IF NOT EXISTS idx_installed_plugins_workspace_id ON installed_plugins(workspace_id);
CREATE INDEX IF NOT EXISTS idx_installed_plugins_source_id ON installed_plugins(source_id);
CREATE INDEX IF NOT EXISTS idx_installed_plugins_version_sha ON installed_plugins(version_sha);
CREATE INDEX IF NOT EXISTS idx_unique_workspace_plugin ON installed_plugins(workspace_id, plugin_identifier);

CREATE TABLE IF NOT EXISTS plugin_privileges (
	plugin_id UUID REFERENCES installed_plugins(id) ON DELETE CASCADE,
	name VARCHAR(64) NOT NULL,
	description TEXT,
	denied_at TIMESTAMPTZ DEFAULT NULL,
	created_at TIMESTAMPTZ DEFAULT now(),
	updated_at TIMESTAMPTZ DEFAULT NULL,
	UNIQUE (plugin_id, name)
);

CREATE INDEX IF NOT EXISTS idx_plugin_privileges_plugin_id ON plugin_privileges(plugin_id);
CREATE INDEX IF NOT EXISTS idx_plugin_privileges_denied_at ON plugin_privileges(denied_at) WHERE denied_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_plugin_privileges_created_at ON plugin_privileges(created_at);
CREATE INDEX IF NOT EXISTS idx_installed_plugins_modes ON installed_plugins USING GIN (modes);
CREATE INDEX IF NOT EXISTS idx_installed_plugins_entry_types ON installed_plugins USING GIN (entry_types);
CREATE INDEX IF NOT EXISTS idx_installed_plugins_metadata ON installed_plugins USING GIN (metadata);

-- Trigger to set the updated_at column
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON installed_plugins
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON plugin_privileges
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

