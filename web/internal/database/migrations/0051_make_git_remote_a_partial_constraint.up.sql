-- Write your migration here
ALTER TABLE plugin_sources
DROP CONSTRAINT IF EXISTS idx_plugin_sources_workspace_id_url;

-- The Git URL and workspace are unique if the git_remote is not null
CREATE UNIQUE INDEX idx_plugin_sources_workspace_id_remote ON plugin_sources(workspace_id, git_remote)
WHERE git_remote IS NOT NULL;

-- The name and workspace are unique if the git_remote is NULL
CREATE UNIQUE INDEX idx_plugin_sources_workspace_id_name ON plugin_sources(workspace_id, name)
WHERE git_remote IS NULL;

