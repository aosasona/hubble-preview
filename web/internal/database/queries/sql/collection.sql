-- name: FindCollectionsByWorkspaceAndUser :many
with
    members_count as (
        select collection_id, count(id) as count
        from collection_members
        group by collection_id
    ),
    entries as (
        select collection_id, count(id) as count from entries group by collection_id
    )
select c.*, coalesce(mc.count, 0) as member_count, coalesce(e.count, 0) as entry_count
from collections c
right join collection_members cm on cm.collection_id = c.id
left join entries e on e.collection_id = c.id
left join members_count mc on mc.collection_id = c.id
where
    cm.user_id = @user_id
    and c.workspace_id = @workspace_id
    and c.deleted_at is null
    and cm.deleted_at is null
order by c.name asc
;

-- name: CollectionNameExists :one
select count(id) > 0 as exists
from collections
where
    workspace_id = @workspace_id
    and lower(name) = lower(@collection_name)
    and deleted_at is null
;

-- name: CollectionSlugExists :one
select count(id) > 0 as exists
from collections
where workspace_id = @workspace_id and slug = @slug and deleted_at is null
;

-- name: CreateCollection :one
insert into collections
(workspace_id, name, description, slug, owner_id)
values (@workspace_id, @name, @description, @slug, @owner_id)
returning *;

-- name: AssignAllWorkspaceMembersToCollection :exec
insert into collection_members (collection_id, user_id, bitmask_role)
select @collection_id, user_id, @role
from workspace_members where
    workspace_members.user_id != @owner_id
    and workspace_id = @workspace_id
    and deleted_at is null
on conflict (collection_id, user_id)
    do update set deleted_at = null, bitmask_role = EXCLUDED.bitmask_role
;

-- name: CreateCollectionMember :exec
-- Create a collection member. If the user already exists in the collection and was
-- deleted, restore them.
insert into collection_members (collection_id, user_id, bitmask_role)
values (@collection_id, @user_id, @bitmask_role)
on conflict (collection_id, user_id)
    do update set deleted_at = null, bitmask_role = EXCLUDED.bitmask_role
;

-- name: FindCollectionIdByPublicId :one
select c.id::integer
from collections c
right join workspaces w on w.id = c.workspace_id
where w.public_id = @workspace_id and c.public_id = @public_id and c.deleted_at is null
;


-- name: FindCollectionMember :one
select sqlc.embed(cm)
from collection_members cm
join collections c on c.id = cm.collection_id
join workspaces w on w.id = c.workspace_id
where
    (c.slug = @collection_slug or c.public_id = @collection_public_id)
    and (w.slug = @workspace_slug or w.public_id = @workspace_public_id)
    and cm.user_id = @user_id
    and c.deleted_at is null
    and cm.deleted_at is null
;

-- name: FindCollectionWithMembershipStatus :one
-- Find a collection by its public_id or slug and check if the user has permissions to
-- access it
select
    sqlc.embed(c),
    sqlc.embed(w),
    (cm.id is not null and cm.deleted_at is null)::bool as is_member,
    -- remove the role if the user has been deleted
    case
        when (cm.id is not null and cm.deleted_at is null) then cm.bitmask_role else 0
    end as role,
    cm.id as member_id,
    cm.user_id as member_user_id
from collections c
left join collection_members cm on cm.collection_id = c.id and cm.user_id = @user_id
left join workspaces w on w.id = c.workspace_id
where
    (w.public_id = @workspace_public_id or w.slug = @workspace_slug)
    and (c.public_id = @collection_public_id or c.slug = @collection_slug)
    and c.deleted_at is null
    and w.deleted_at is null
;

-- name: FindCollectionMembers :many
with
    members as (
        select
            cm.id,
            u.id as user_id,
            u.public_id as public_user_id,
            u.first_name,
            u.last_name,
            u.email,
            cm.bitmask_role as role,
            cm.created_at,
            'accepted' as status,
            '' as invite_id
        from collection_members cm
        join users u on u.id = cm.user_id
        join collections c on c.id = cm.collection_id
        join workspaces w on w.id = c.workspace_id
        where
            (w.public_id = @workspace_id or w.slug = @workspace_slug)
            and (c.public_id = @collection_id or c.slug = @collection_slug)
            and cm.deleted_at is null
            and w.deleted_at is null
            and u.deleted_at is null
            and c.deleted_at is null
    )
select *, count(*) over () as total_count
from members
limit $1
offset $2
;


-- name: FindUserCollectionsWithWorkspaceOwner :many
-- Find all collections that the target user owns and thw workspace owner is in too
with
    -- get all the target user's owned collections
    user_collections as (
        select c.id, c.workspace_id, c.owner_id
        from collections c
        join collection_members cm on cm.collection_id = c.id
        where
            cm.user_id = @user_id
            and c.workspace_id = @workspace_id
            and c.owner_id = cm.user_id
            and c.deleted_at is null
    )
select distinct u.id
from user_collections u
join collection_members cm on u.id = cm.collection_id
join workspaces w on u.workspace_id = w.id
where w.owner_id = cm.user_id and cm.deleted_at is null and w.deleted_at is null
;

-- name: ReassignCollectionOwnership :exec
-- Reassign the ownership of a collection to another user
-- Also update the collection members to reflect the new owner
with
    target_collections as (
        update collection_members
        set bitmask_role = @owner_role,
        deleted_at = null
        where
            collection_id = any(@collection_ids::integer[]) and user_id = @new_owner_id
        returning collection_id
    )
    update collections
    set owner_id = @new_owner_id
from target_collections
where
    collections.id = target_collections.collection_id
    and collections.owner_id = @old_owner_id
    and collections.workspace_id = @workspace_id
    and collections.deleted_at is null
returning id
;

-- name: MarkUsersCollectionsAsDeleted :exec
-- Mark all collections that the user owns as deleted
update collections
set deleted_at = now()
where owner_id = @user_id
    and workspace_id = @workspace_id
    and deleted_at is null
;

-- name: MarkCollectionAsDeleted :exec
-- Mark a collection as deleted
update collections
set deleted_at = now()
where
	(id = @collection_id or public_id = @collection_public_id)
	and deleted_at is null
;

-- name: UpdateCollectionDetails :one
update collections
set name = case
        when sqlc.narg('name')::varchar is not null then @name::varchar
        else name
    end,
    description = case
        when sqlc.narg('description')::text is not null then @description::text
        else description
    end,
    slug = case
        when sqlc.narg('slug')::text is not null then @slug::text
        else slug
    END
where id = @collection_id and deleted_at is null
returning *;


-- name: AddMembersToCollection :many
-- Add members to a collection from the workspace members list using their email
-- address. IF they are already in the collection, update their role and restore them
insert into collection_members (collection_id, user_id, bitmask_role)
select c.id, u.id, @user_role
from collections c
join workspace_members wm ON wm.workspace_id = c.workspace_id
join users u ON u.id = wm.user_id -- this is here for further validation (e.g. that the user has not been deleted)
	where (c.id = @collection_id or c.public_id = @collection_public_id)
		and wm.user_id != c.owner_id -- don't add owner
		and u.email = any(@emails::text[])
		and wm.deleted_at is null
		and c.deleted_at is null
		and u.deleted_at is null
on conflict (collection_id, user_id)
	do update set
		deleted_at = null,
		bitmask_role = EXCLUDED.bitmask_role
	where collection_members.deleted_at is not null -- only restore the ones that were deleted instead of accidentally overwriting roles
returning *;


-- name: RemoveCollectionMembers :many
-- Remove members from a collection by their email address.
with
    targets as (
        select cm.id
        from collection_members cm
        join users u on u.id = cm.user_id
        join collections c on c.id = cm.collection_id
        where
            c.id = @collection_id
            and c.workspace_id = @workspace_id
            and cm.user_id != c.owner_id
            and (
                sqlc.arg('include_admins')::bool = true
                or cm.bitmask_role & @admin_role = 0
            )
            and u.email = any(@emails::text[])
            and cm.deleted_at is null
            and c.deleted_at is null
    )
    update collection_members
    set deleted_at = now()
where
    id in (select id from targets)
    and collection_members.deleted_at is null
    and collection_members.collection_id = @collection_id
returning id
;

-- name: LeaveCollection :exec
-- Remove the user from the collection
update collection_members
set deleted_at = now()
where collection_id = @collection_id
    and user_id = @user_id
    and deleted_at is null
;

