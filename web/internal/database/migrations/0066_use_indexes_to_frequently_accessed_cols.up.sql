-- Users
CREATE INDEX IF NOT EXISTS idx_users_public_id ON users (public_id);
CREATE INDEX IF NOT EXISTS idx_users_email_address ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);

-- Collections
CREATE INDEX IF NOT EXISTS idx_collections_public_id ON collections (public_id);
CREATE INDEX IF NOT EXISTS idx_collections_slug ON collections (slug);
CREATE INDEX IF NOT EXISTS idx_collections_owner_id ON collections (owner_id);
CREATE INDEX IF NOT EXISTS idx_collections_deleted_at ON collections (deleted_at);
CREATE INDEX IF NOT EXISTS idx_collections_slug_deleted_at ON collections (slug, deleted_at);
CREATE INDEX IF NOT EXISTS idx_collections_pbid_deleted_at ON collections (public_id, deleted_at);

-- Collection members
CREATE INDEX IF NOT EXISTS idx_collection_members_cid ON collection_members (collection_id);
CREATE INDEX IF NOT EXISTS idx_collection_members_uid ON collection_members (user_id);
CREATE INDEX IF NOT EXISTS idx_collection_members_deleted_at ON collection_members (deleted_at);
CREATE INDEX IF NOT EXISTS idx_collection_members_cid_user ON collection_members (collection_id, user_id);
CREATE INDEX IF NOT EXISTS idx_collection_members_role ON collection_members (bitmask_role);

-- Workspaces members
CREATE INDEX IF NOT EXISTS idx_workspace_members_wid ON workspace_members (workspace_id);
CREATE INDEX IF NOT EXISTS idx_workspace_members_uid ON workspace_members (user_id);
CREATE INDEX IF NOT EXISTS idx_workspace_members_deleted_at ON workspace_members (deleted_at);
CREATE INDEX IF NOT EXISTS idx_workspace_members_wid_user ON workspace_members (workspace_id, user_id);
CREATE INDEX IF NOT EXISTS idx_workspace_members_role ON workspace_members (bitmask_role);

-- Workspaces
CREATE INDEX IF NOT EXISTS idx_workspaces_public_id ON workspaces (public_id);
CREATE INDEX IF NOT EXISTS idx_workspaces_slug ON workspaces (slug);
CREATE INDEX IF NOT EXISTS idx_workspaces_owner_id ON workspaces (owner_id);
CREATE INDEX IF NOT EXISTS idx_workspaces_deleted_at ON workspaces (deleted_at);
CREATE INDEX IF NOT EXISTS idx_workspaces_slug_deleted_at ON workspaces (slug, deleted_at);
CREATE INDEX IF NOT EXISTS idx_workspaces_pbid_deleted_at ON workspaces (public_id, deleted_at);


-- Entries
CREATE INDEX IF NOT EXISTS idx_entries_file_id ON entries (file_id);
CREATE INDEX IF NOT EXISTS idx_entries_parent_id ON "entries" (parent_id);
CREATE INDEX IF NOT EXISTS idx_entries_collection_id ON entries (collection_id);
CREATE INDEX IF NOT EXISTS idx_entries_created_at ON entries (created_at);
CREATE INDEX IF NOT EXISTS idx_entries_updated_at ON entries (updated_at);
CREATE INDEX IF NOT EXISTS idx_entries_public_id ON entries (public_id);

-- Entries queue
CREATE INDEX IF NOT EXISTS idx_entries_queue_entry_id ON entries_queue (entry_id);
CREATE INDEX IF NOT EXISTS idx_entries_queue_status ON entries_queue (status);
CREATE INDEX IF NOT EXISTS idx_entries_queue_created_at ON entries_queue (created_at);
CREATE INDEX IF NOT EXISTS idx_entries_queue_updated_at ON entries_queue (updated_at);
CREATE INDEX IF NOT EXISTS idx_entries_queue_attempts ON entries_queue (attempts);

-- Entries chunks
CREATE INDEX IF NOT EXISTS idx_entry_chunks_chk_idx ON entry_chunks (chunk_index);
CREATE INDEX IF NOT EXISTS idx_entry_chunks_entry_id ON entry_chunks (entry_id);
CREATE INDEX IF NOT EXISTS idx_entry_chunks_created_at ON entry_chunks (created_at);
CREATE INDEX IF NOT EXISTS idx_entry_chunks_updated_at ON entry_chunks (updated_at);
CREATE INDEX IF NOT EXISTS idx_entry_chunks_embedding_status ON entry_chunks (embedding_status);
CREATE INDEX IF NOT EXISTS idx_entry_chunks_embedding_status_updated_at ON entry_chunks (embedding_status_updated_at);

-- Plugin sources
CREATE INDEX IF NOT EXISTS idx_plugin_sources_metadata ON plugin_sources USING GIN(metadata);
CREATE INDEX IF NOT EXISTS idx_plugin_sources_url ON plugin_sources (git_remote);
CREATE INDEX IF NOT EXISTS idx_plugin_sources_wkid ON plugin_sources (workspace_id);

-- Plugins
CREATE INDEX IF NOT EXISTS idx_plugins_metadata ON installed_plugins USING GIN(metadata);
CREATE INDEX IF NOT EXISTS idx_plugins_plugin_identifier ON installed_plugins (plugin_identifier);
CREATE INDEX IF NOT EXISTS idx_plugin_workspace_id ON installed_plugins (workspace_id);
CREATE INDEX IF NOT EXISTS idx_plugin_version_sha ON installed_plugins (version_sha);

-- MFA accounts
CREATE INDEX IF NOT EXISTS idx_mfa_accounts_user_id ON mfa_accounts (user_id);
CREATE INDEX IF NOT EXISTS idx_mfa_accounts_created_at ON mfa_accounts (created_at);
CREATE INDEX IF NOT EXISTS idx_mfa_accounts_updated_at ON mfa_accounts (updated_at);

-- MFA backup tokens
CREATE INDEX IF NOT EXISTS idx_mfa_backup_tokens_user_id ON mfa_backup_tokens (user_id);

-- TOTP secrets
CREATE INDEX IF NOT EXISTS idx_totp_secrets_account_id ON totp_secrets (account_id);
CREATE INDEX IF NOT EXISTS idx_totp_secrets_version ON totp_secrets (version);

-- Workspace invites
CREATE INDEX IF NOT EXISTS idx_workspace_invites_invite_id ON workspace_invites (invite_id);
CREATE INDEX IF NOT EXISTS idx_workspace_invites_workspace_id ON workspace_invites (workspace_id);
CREATE INDEX IF NOT EXISTS idx_workspace_invites_email ON workspace_invites (email);
CREATE INDEX IF NOT EXISTS idx_workspace_invites_accepted_at ON workspace_invites (accepted_at);

