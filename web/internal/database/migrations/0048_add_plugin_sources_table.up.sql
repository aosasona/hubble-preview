CREATE TYPE versioning_strategy AS ENUM ('tag', 'commit');

CREATE TYPE plugin_source_type AS ENUM ('git', 'local');

CREATE TYPE plugin_source_auth_method AS ENUM ('none', 'ssh', 'https');

CREATE TYPE plugin_sync_status AS ENUM ('idle', 'syncing', 'error');

CREATE TABLE IF NOT EXISTS plugin_sources (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	workspace_id integer REFERENCES workspaces(id) ON DELETE CASCADE,

	name TEXT NOT NULL,
	description TEXT,
	author TEXT NOT NULL,
	disabled_at TIMESTAMPTZ DEFAULT NULL,
	versioning_strategy versioning_strategy NOT NULL,
	git_remote TEXT,
	auth_method plugin_source_auth_method NOT NULL DEFAULT 'none',
	version_id VARCHAR(64) DEFAULT NULL,

	sync_status plugin_sync_status NOT NULL DEFAULT 'idle',
	last_sync_error TEXT DEFAULT NULL,
	last_synced_at TIMESTAMPTZ DEFAULT NULL,

	added_at TIMESTAMPTZ DEFAULT now(),
	updated_at TIMESTAMPTZ DEFAULT NULL,
	metadata JSONB DEFAULT '{}'::jsonb,

	UNIQUE (workspace_id, name)
);

COMMENT ON COLUMN plugin_sources.workspace_id IS 'The workspace that this plugin source was added to. This is used to determine the scope of the plugin source.';

COMMENT ON COLUMN plugin_sources.metadata IS 'This contains things like the (encrypted) SSH key or HTTP credentials.';

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_plugin_sources_workspace_id ON plugin_sources(workspace_id);
CREATE INDEX IF NOT EXISTS idx_plugin_sources_enabled ON plugin_sources(disabled_at) WHERE disabled_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_plugin_sources_git_remote ON plugin_sources(git_remote);
CREATE INDEX IF NOT EXISTS idx_plugin_sources_sync_status ON plugin_sources(sync_status);

-- Trigger to set the updated_at column
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON plugin_sources
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

