-- name: CreateRemotePluginSource :one
INSERT INTO plugin_sources
	(workspace_id, name, description, author, versioning_strategy, git_remote, auth_method, version_id, sync_status)
SELECT workspaces.id, @name, @description, @author, @versioning_strategy, @git_remote, @auth_method, @version_id, 'idle' FROM workspaces WHERE
	workspaces.id = @workspace_id::integer AND workspaces.deleted_at IS NULL
RETURNING *;

-- name: FindSourceByGitRemote :one
select *
from plugin_sources
where workspace_id = @workspace_id and git_remote = @git_remote
;

-- name: FindPluginSourceByID :one
select *
from plugin_sources
where workspace_id = @workspace_id and id = @source_id
;

-- name: FindPluginSourcesByWorkspaceID :many
with
    all_sources as (
        select p.*
        from plugin_sources p
        where p.workspace_id = @workspace_id and disabled_at is null
    )
select *, count(*) over () as total_count
from all_sources
order by name asc
limit $1
offset $2
;

-- name: SourceExists :one
select count(id) > 0
from plugin_sources
where workspace_id = @workspace_id and git_remote = @git_remote
;

-- name: RemovePluginSource :exec
delete from plugin_sources
where
    workspace_id = @workspace_id
    and (
        id = @source_id
        or (sqlc.narg('git_remote')::text is not null and git_remote = @git_remote)
    )
;

-- name: RemoveSourcePlugins :exec
with
    source as (
        select id
        from plugin_sources
        where workspace_id = @workspace_id and git_remote = @git_remote
    )
delete from installed_plugins
where
    installed_plugins.workspace_id = @workspace_id
    and source_id in (select id from source)
;


-- name: FindInstalledPluginsByWorkspaceID :many
select
    sqlc.embed(i), s.id as source_id, s.name as source_name, s.git_remote as source_url
from installed_plugins i
join plugin_sources s on s.id = i.source_id
where s.workspace_id = @workspace_id
;

-- name: FindInstalledPlugin :one
select
    sqlc.embed(i), s.id as source_id, s.name as source_name, s.git_remote as source_url
from installed_plugins i
join plugin_sources s on s.id = i.source_id
where s.workspace_id = @workspace_id and i.plugin_identifier = @plugin_identifier
;

-- name: UpsertInstalledPlugin :one
insert into installed_plugins (
    plugin_identifier,
    workspace_id,
    source_id,
    name,
    description,
    modes,
    entry_types,
    version_sha,
    last_updated_at,
    privileges,
    added_at
) values(
    @identifier,
    @workspace_id,
    @source_id,
    @name,
    @description,
    @modes,
    @targets,
    @checksum,
    @plugin_last_updated_at,
    @privileges,
    now()
) on conflict (workspace_id, plugin_identifier) do update set
    name = excluded.name,
    description = excluded.description,
    modes = excluded.modes,
    entry_types = excluded.entry_types,
    version_sha = excluded.version_sha,
    privileges = excluded.privileges,
    last_updated_at = excluded.last_updated_at
returning *;

-- name: RemoveInstalledPlugin :exec
delete from installed_plugins
where workspace_id = @workspace_id and plugin_identifier = @plugin_identifier
;

-- name: PluginKvGet :one
select value
from plugins_kv
where key = @key and plugin_id = @identifier
;

-- name: PluginKvSet :one
insert into plugins_kv (plugin_id, key, value)
values (@identifier, @key, @value)
on conflict (plugin_id, key) do update set
	value = excluded.value
returning *;

-- name: PluginKvDelete :exec
delete from plugins_kv
where plugin_id = @identifier and key = @key
;

-- name: PluginKvDeleteByPluginID :exec
delete from plugins_kv
where plugin_id = @identifier
;

-- name: PluginKvGetByPluginID :many
select key, value
from plugins_kv
where plugin_id = @identifier
;


-- name: FindOnCreatePluginsForType :many
-- This finds all plugins that support the given entry type and have the `on_create`
-- mode
select sqlc.embed(i)
from installed_plugins i
join entries e on e.id = @entry_id
join workspaces w on w.id = i.workspace_id
where
    'on_create' = any(i.modes)
    and e.entry_type = any(i.entry_types)
    and w.public_id = @workspace_public_id
    and w.deleted_at is null
;

