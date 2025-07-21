-- name: GetUserMembership :one
select *
from workspace_members
where workspace_id = $1 and user_id = $2 and deleted_at is null
;

-- name: FindWorkspaceIdBySlug :one
select public_id
from workspaces
where slug = @slug and deleted_at is null
;

-- name: FindWorkspaceByPublicID :one
select *
from workspaces
where public_id = $1 and deleted_at is null
;

-- name: FindWorkspaceByID :one
select sqlc.embed(w)
from workspaces w
where id = $1 and deleted_at is null
;

-- name: FindAllWorkspacesByUserID :many
select w.*
from workspaces w
inner join workspace_members wm on wm.workspace_id = w.id
where wm.user_id = $1 and w.deleted_at is null and wm.deleted_at is null
order by w.display_name asc, w.created_at desc
;


-- name: FindWorkspaceWithMembershipStatus :one
-- Find a workspace by its public_id or slug and check if the user has permissions to
-- access it
select
    w.*,
    (wm.id is not null and wm.deleted_at is null)::bool as is_member,
    case
        when (wm.id is not null and wm.deleted_at is null) then wm.bitmask_role else 0
    end as role,
    wm.id as member_id,
    wm.user_id as member_user_id
from workspaces w
left join workspace_members wm on wm.workspace_id = w.id and wm.user_id = @user_id
where (w.public_id = @public_id or w.slug = @slug) and w.deleted_at is null
;


-- name: CreateWorkspace :one
insert into workspaces (display_name, owner_id, description, slug)
values (@name, @owner_id, @description, @slug)
returning *;

-- name: WorkspaceNameExists :one
select count(id) > 0 as exists
from workspaces
where
    lower(display_name) = lower(@display_name)
    and owner_id = @owner_id
    and deleted_at is null
;

-- name: WorkspaceSlugExists :one
select count(id) > 0 as exists
from workspaces
where slug = @slug
;

-- name: WorkspaceMemberExists :one
select (count(wm.id) > 0) as exists
from workspace_members wm
join users u on u.id = wm.user_id
where
    wm.workspace_id = @workspace_id
    and (sqlc.narg('email')::text is null or u.email = sqlc.narg('email')::text)
    and (sqlc.narg('user_id')::integer is null or u.id = sqlc.narg('user_id')::integer)
    and (
        sqlc.narg('public_user_id')::uuid is null
        or u.public_id = sqlc.narg('public_user_id')::uuid
    )
    and wm.deleted_at is null
;

-- name: CollectionExistsInWorkspace :one
select count(c.id) > 0 as exists
from collections c
left join workspaces w on w.id = c.workspace_id
where w.public_id = @workspace_id and c.public_id = @collection_id
;

-- name: CreateWorkspaceMember :one
-- Create a workspace member. If the user already exists in the workspace and was
-- deleted, restore them.
insert into workspace_members(workspace_id, user_id, bitmask_role)
values(@workspace_id, @user_id, @role)
on conflict (workspace_id, user_id) do update set
    bitmask_role = EXCLUDED.bitmask_role,
    deleted_at = null
returning *;

-- name: UpsertWorkspaceInvite :one
insert into workspace_invites (workspace_id, email, role, invited_by)
values (@workspace_id, @email, @role, @invited_by)
-- reset the invite if it already exists
on conflict (workspace_id, email) do update set
	invite_id = gen_random_uuid(),
    role = EXCLUDED.role,
	invited_by = EXCLUDED.invited_by,
    invited_at = now(),
    accepted_at = null,
    declined_at = null,
    deleted_at = null
returning *;

-- name: FindWorkspaceMembers :many
with
    members_and_invites as (
        select
            wm.id,
            u.id as user_id,
            u.public_id as public_user_id,
            u.first_name,
            u.last_name,
            u.email,
            wm.bitmask_role as role,
            wm.created_at,
            'accepted' as status,
            '' as invite_id
        from workspace_members wm
        left join users u on u.id = wm.user_id
        left join workspaces w on w.id = wm.workspace_id
        where w.public_id = @workspace_id and wm.deleted_at is null
        union all
        select
            i.id,
            coalesce(u.id, null) as user_id,
            coalesce(u.public_id, null) as public_user_id,
            coalesce(u.first_name, null) as first_name,
            coalesce(u.last_name, null) as last_name,
            i.email,
            i.role,
            i.invited_at,
            -- fmt: off
            case
                when i.declined_at is not null then 'declined'
                when i.deleted_at is not null then 'revoked'
                when i.invited_at + interval '14 days' < now() then 'expired'
                else 'pending'
            end as status,
            i.invite_id::text as invite_id
        from workspace_invites i
        left join workspaces w on w.id = i.workspace_id
        left join users u on u.email = i.email
        where
            w.public_id = @workspace_id
            and i.accepted_at is null
            and i.deleted_at is null
    )
select *, count(*) over () as total_count
from members_and_invites
limit $1
offset $2
;

-- name: FindInviteById :one
select
    i.id,
    i.invite_id,
    w.id as workspace_id,
    w.public_id as workspace_public_id,
    w.display_name as workspace_name,
    w.slug as workspace_slug,
    sqlc.embed(inviter),
    i.role,
    -- fmt: off
    case
        when i.accepted_at is not null then 'accepted'
        when i.declined_at is not null then 'declined'
        when i.deleted_at is not null then 'revoked'
        when i.invited_at + interval '14 days' < now() then 'expired'
        else 'pending'
    end as status,
    i.email as invited_email,
    i.invited_at,
    u.id as invited_user_id,
    (u.id is not null)::bool as invited_user_exists
from workspace_invites i
join workspaces w on w.id = i.workspace_id
join users inviter on inviter.id = i.invited_by
left join users u on u.email = i.email
where i.invite_id = @invite_id and inviter.deleted_at is null and w.deleted_at is null
;

-- name: FindWorkspaceMemeber :one
select
    wm.id,
    u.id as user_id,
    u.public_id as public_user_id,
    u.first_name,
    u.last_name,
    u.email,
    wm.bitmask_role as role,
    wm.created_at,
    -- fmt: off
    case
        when i.accepted_at is not null then 'accepted'
        when i.declined_at is not null then 'declined'
        when i.deleted_at is not null then 'revoked'
        when i.invited_at + interval '14 days' < now() then 'expired'
        else 'pending'
    end as status,
    i.invite_id as invite_id
from workspace_members wm
join users u on u.id = wm.user_id
join workspaces w on w.id = wm.workspace_id
left join workspace_invites i on i.email = u.email and i.workspace_id = w.id
where w.public_id = @workspace_id
    and (sqlc.narg('user_id')::integer is null or wm.user_id = sqlc.narg('user_id')::integer)
    and (sqlc.narg('email')::text is null or u.email = sqlc.narg('email')::text)
    and (sqlc.narg('member_id')::integer is null or wm.id = sqlc.narg('member_id')::integer)
    and wm.deleted_at is null
	and w.deleted_at is null
	and u.deleted_at is null
;


-- name: UpdateWorkspaceInviteStatus :one
update workspace_invites
set accepted_at = case when @status::text = 'accepted' then now() else null end,
	declined_at = case when @status::text = 'declined' then now() else null end,
	deleted_at = case when @status::text = 'revoked' then now() else null end
where invite_id = @invite_id
returning *;

-- name: UpdateWorkspaceMemberRole :exec
update workspace_members
set bitmask_role = @role
from users
where workspace_members.user_id = users.id
    and workspace_members.workspace_id = @workspace_id
    and users.public_id = @user_id
	and workspace_members.deleted_at is null
	and users.deleted_at is null
;

-- name: UpdateAllCollectionMemberRole :exec
-- Update all collection members roles where the user is a member but not an owner
update collection_members
set bitmask_role = @role
from users
where
    collection_members.user_id = users.id
    and users.public_id = @user_id
    and (collection_members.bitmask_role & @owner_role) = 0
;

-- name: MarkWorkspaceMemberAsDeleted :execresult
-- Mark a workspace member as deleted. This is used when a user leaves a workspace
-- or is removed from it.
update workspace_members
set deleted_at = now()
where workspace_id = @workspace_id and user_id = @user_id;


-- name: UpdateWorkspaceDetails :one
update workspaces
set display_name = case
        when sqlc.narg('name')::varchar is not null then @name::varchar
        else display_name
    end,
    description = case
        when sqlc.narg('description')::text is not null then @description::text
        else description
    end,
    slug = case
        when sqlc.narg('slug')::text is not null then @slug::text
        else slug
    end
where id = @workspace_id and deleted_at is null
returning *;


-- name: MarkWorkspaceAsDeleted :execresult
-- Mark a workspace as deleted, and all its entries, collections, and members as deleted.
update workspaces
set deleted_at = now()
where public_id = @workspace_id and deleted_at is null;

