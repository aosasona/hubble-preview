-- Drop triggers
DROP TRIGGER IF EXISTS set_updated_at ON plugin_privileges;
DROP TRIGGER IF EXISTS set_updated_at ON installed_plugins;

-- Drop indexes
DROP INDEX IF EXISTS idx_installed_plugins_metadata, idx_installed_plugins_entry_types, idx_installed_plugins_modes, idx_plugin_privileges_created_at, idx_plugin_privileges_denied_at, idx_plugin_privileges_plugin_id;

-- Drop plugin_privileges table
DROP TABLE IF EXISTS plugin_privileges;

-- Drop indexes on installed_plugins
DROP INDEX IF EXISTS idx_installed_plugins_version_sha, idx_installed_plugins_source_id, idx_installed_plugins_workspace_id, idx_unique_workspace_plugin;

-- Drop installed_plugins table
DROP TABLE IF EXISTS installed_plugins;

-- Drop enum types
DROP TYPE IF EXISTS plugin_mode;
DROP TYPE IF EXISTS plugin_scope;

