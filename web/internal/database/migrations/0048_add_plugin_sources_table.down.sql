-- Drop triggers
DROP TRIGGER IF EXISTS set_updated_at ON plugin_sources;

-- Drop indexes
DROP INDEX IF EXISTS idx_plugin_sources_sync_status, idx_plugin_sources_git_remote, idx_plugin_sources_enabled, idx_plugin_sources_workspace_id;

-- Drop plugin_sources table
DROP TABLE IF EXISTS plugin_sources;

DROP TYPE IF EXISTS versioning_strategy;
DROP TYPE IF EXISTS plugin_source_type;
DROP TYPE IF EXISTS plugin_source_auth_method;
DROP TYPE IF EXISTS  plugin_sync_status;

