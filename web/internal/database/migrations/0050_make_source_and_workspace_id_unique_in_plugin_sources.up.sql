ALTER TABLE plugin_sources
DROP CONSTRAINT IF EXISTS plugin_sources_workspace_id_name_key,
ADD CONSTRAINT idx_plugin_sources_workspace_id_url UNIQUE (workspace_id, git_remote);

