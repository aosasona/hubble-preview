-- name: CreateLinkEntry :one
insert into entries (name, meta, version, entry_type, collection_id, added_by, last_updated_by) values (@name, @meta, 1, 'link', @collection_id, @user_id, @user_id) returning *;

-- name: CreateFileEntry :one
insert into entries (name, meta, content, file_id, entry_type, checksum, collection_id, added_by, last_updated_by, filesize_bytes) values (@name, @meta, @content, @file_id, @entry_type, @checksum, @collection_id, @added_by, @added_by, @filesize_bytes) returning *;

-- name: UpdateEntry :one
update entries
set name = case when sqlc.narg('name')::text is null then name else @name end,
	content = case when sqlc.narg('content')::text is null then content else @content end,
	text_content = case when sqlc.narg('text_content')::text is null then text_content else @text_content end,
	checksum = case when sqlc.narg('checksum')::text is null then checksum else @checksum end
where public_id = @public_id and deleted_at is null
returning *;

-- name: FindEntryInCollectionAndWorkspace :one
select
    sqlc.embed(e),
    sqlc.embed(c),
    u.first_name as added_by_first_name,
    u.last_name as added_by_last_name,
    u.username as added_by_username,
    q.status,
    q.created_at as queued_at
from entries e
join collections c on c.id = e.collection_id
join workspaces w on w.id = c.workspace_id
join entries_queue q on q.entry_id = e.id
join users u on u.id = e.added_by
where
    e.public_id = @entry_public_id
    and w.public_id = @workspace_id
    and c.public_id = @collection_id
    and e.deleted_at is null
    and c.deleted_at is null
    and w.deleted_at is null
;

-- name: FindEntryById :one
select
    e.id,
    e.public_id,
    e.parent_id,
    e.origin,
    e.content,
    e.text_content,
    e.name,
    e.meta,
    e.version,
    e.entry_type as type,
    e.file_id,
    e.filesize_bytes,
    e.added_by,
    e.last_updated_by,
    e.created_at,
    e.updated_at,
    e.archived_at,
    c.name as collection_name,
    c.public_id as collection_id,
    c.slug as collection_slug,
    u.first_name as added_by_first_name,
    u.last_name as added_by_last_name,
    u.username as added_by_username,
    u.avatar_id as added_by_avatar_id,
    w.public_id as workspace_id,
    w.display_name as workspace_name,
    w.slug as workspace_slug,
    q.status,
    q.created_at as queued_at
from entries e
join collections c on c.id = e.collection_id
join workspaces w on w.id = c.workspace_id
join entries_queue q on q.entry_id = e.id
join users u on u.id = e.added_by
where
    (
        (sqlc.narg('entry_public_id')::uuid is null and e.id = @entry_id)
        or (sqlc.narg('entry_id')::integer is null and e.public_id = @entry_public_id)
    )
    and (sqlc.narg('workspace_slug')::text is null or w.slug = @workspace_slug)
    and (
        sqlc.narg('workspace_public_id')::uuid is null
        or w.public_id = @workspace_public_id
    )
    and (sqlc.narg('collection_slug')::text is null or c.slug = @collection_slug)
    and (
        sqlc.narg('collection_public_id')::uuid is null
        or c.public_id = @collection_public_id
    )
    and e.deleted_at is null
    and c.deleted_at is null
    and w.deleted_at is null
limit 1
;

-- name: FindEntries :many
with
    latest_entries as (
        select
            e.*,
            row_number() over (
                partition by coalesce(e.parent_id, e.id) order by e.version desc
            ) as rn
        from entries e
        join collections c on c.id = e.collection_id
        join workspaces w on w.id = c.workspace_id
        join collection_members cm on cm.collection_id = c.id
        join workspace_members wm on wm.workspace_id = w.id
        where
            cm.user_id = @user_id
            and wm.user_id = @user_id
            and e.deleted_at is null
            and e.archived_at is null
            and (
                sqlc.narg('workspace_slug')::text is null
                or w.slug = sqlc.narg('workspace_slug')::text
            )
            and (
                sqlc.narg('workspace_public_id')::uuid is null
                or w.public_id = sqlc.narg('workspace_public_id')::uuid
            )
            and (
                sqlc.narg('collection_slug')::text is null
                or c.slug = sqlc.narg('collection_slug')::text
            )
            and (
                sqlc.narg('collection_public_id')::uuid is null
                or c.public_id = sqlc.narg('collection_public_id')::uuid
            )
            and c.deleted_at is null
            and w.deleted_at is null
            and e.deleted_at is null
            and e.archived_at is null
        order by e.collection_id, e.version desc
    )
select
    e.id,
    e.public_id,
    e.parent_id,
    e.origin,
    e.content,
    e.text_content,
    e.name,
    e.meta,
    e.version,
    e.entry_type as type,
    e.file_id,
    e.filesize_bytes,
    e.added_by,
    e.last_updated_by,
    e.created_at,
    e.updated_at,
    e.archived_at,
    c.name as collection_name,
    c.public_id as collection_id,
    c.slug as collection_slug,
    u.first_name as added_by_first_name,
    u.last_name as added_by_last_name,
    u.username as added_by_username,
    u.avatar_id as added_by_avatar_id,
    w.public_id as workspace_id,
    w.display_name as workspace_name,
    w.slug as workspace_slug,
    q.status,
    q.created_at as queued_at,
    count(*) over () as total_entries
from latest_entries e
join collections c on c.id = e.collection_id
join workspaces w on w.id = c.workspace_id
join entries_queue q on q.entry_id = e.id
join users u on u.id = e.added_by
order by coalesce(e.updated_at, e.created_at) desc, q.updated_at desc
limit $1
offset $2
;

-- name: GetEntriesOwnership :many
select
    e.public_id,
    e.added_by,
    (e.added_by = cm.user_id) as is_owner,
    cm.bitmask_role as user_role
from entries e
join collection_members cm on cm.collection_id = e.collection_id
where cm.user_id = @user_id and e.public_id = any(@entry_public_ids::uuid[])
;

-- name: DeleteEntries :many
delete from entries
where public_id = any(@entry_public_ids::uuid[]) and deleted_at is null
returning public_id
;

-- name: EnqueueEntries :copyfrom
insert into entries_queue(entry_id, payload)
values (@entry_id, @payload)
;

-- name: DequeueEntries :many
delete from entries_queue eq
using entries e
where eq.entry_id = e.id and e.public_id = any(@entry_public_ids::uuid[])
returning eq.entry_id
;

-- name: ResolveEntryIds :many
select e.id, e.public_id
from entries e
where e.public_id = any(@entry_public_ids::uuid[]) and e.deleted_at is null
;

-- name: ResolveEntryIdsWithoutChunks :many
select e.id, count(ec.id) as chunk_count
from entries e
left join entry_chunks ec on ec.entry_id = e.id
group by e.id
having
    e.public_id = any(@entry_public_ids::uuid[])
    and e.deleted_at is null
    and count(ec.id) = 0
;


-- name: InsertChunk :one
with
    entry_id as (
        select id from entries where public_id = @entry_public_id and deleted_at is null
    )
    insert into entry_chunks(entry_id, chunk_index, min_version, content, language)
select id, @index, @min_version, @content, @language
from entry_id
returning *
;

-- name: InsertChunks :copyfrom
insert into entry_chunks(entry_id, chunk_index, min_version, content, language)
values (@entry_id, @index, @min_version, @content, @language)
;


-- name: UpdateChunk :one
update entry_chunks
set
    content = case when sqlc.narg('content')::text is null then content else @content end,
    language = case when sqlc.narg('language')::text is null then language else @language end
where entry_id = @entry_id and chunk_index = @index and min_version = @version
returning *;


-- name: FindAllQueuedEntries :many
select e.id
from entries_queue q
join entries e on e.id = q.entry_id
where
    q.attempts <= q.max_attempts
    and (
        q.status in ('queued', 'failed')  -- either it hasn't been processed or it failed
        or (q.status = 'processing' and q.updated_at < now() - interval '12 hours')  -- or it has been marked as processing but hasn't been updated in the last 12 hours
    )
    and q.available_at <= now()
    and e.deleted_at is null
;

-- name: UpdateEntryStatus :exec
update entries_queue
set status = @status,
    attempts = case
        when status = 'queued' then 1
        when status = 'processing' then attempts
        when status = 'failed' and attempts < max_attempts then attempts + 1
        when status = 'failed' and attempts >= max_attempts then 0
        when status = 'completed' or status = 'canceled' then attempts
        else attempts + 1
    end,
    updated_at = now()
where entry_id = @entry_id;

-- name: DeleteEntryChunksByPublicId :many
with
    targets as (
        select e.id, e.version
        from entries e
        join collections c on c.id = e.collection_id
        join workspaces w on w.id = c.workspace_id
        where
            e.public_id = any(@entry_public_ids::uuid[])
            and w.public_id = @workspace_id
            and e.deleted_at is null
    )
delete from entry_chunks ec
using targets t
where ec.entry_id = t.id and ec.min_version >= t.version
returning ec.entry_id::integer
;

-- name: RequeueEntries :many
update entries_queue
set status = 'queued', attempts = 0, available_at = now()
where entry_id = any(@entry_ids::integer[])
returning entry_id
;

-- name: FindEntryChunks :many
select ec.id, ec.content, ec.embedding_status
from entry_chunks ec
join entries e on e.id = ec.entry_id
where
    e.public_id = @entry_public_id
    and (
        ec.embedding_status = 'pending'
        -- we will only retry failed chunks that have failed less than 5 times
        or (ec.embedding_status = 'failed' and ec.embedding_error_count < 5)
        or (
            -- this is a special case where the chunk was marked as processing but
            -- hasn't
            -- been updated in the last 2 hours
            ec.embedding_status = 'processing'
            and ec.embedding_status_updated_at < now() - interval '2 hours'
        )
    )
    and ec.semantic_vector is null
    and ec.content is not null
    and ec.content != ''
    and e.deleted_at is null
;

-- name: FindUnindexedChunks :many
select id, content
from entry_chunks
where
    embedding_status = 'pending'
    -- we will only retry failed chunks that have failed less than 5 times
    or (embedding_status = 'failed' and embedding_error_count < 5)
    or (
        -- this is a special case where the chunk was marked as processing but hasn't
        -- been updated in the last 2 hours
        embedding_status = 'processing'
        and embedding_status_updated_at < now() - interval '2 hours'
    )
    and semantic_vector is null
    and content is not null
    and content != ''
;

-- name: ChunkCanBeProcessed :one
-- the actual data is passed to the handlers in the queue, but to prevent duplicate
-- jobs, handlers need to ensure they can actually process what they have just gotten
select count(*) > 0 as can_process
from entry_chunks ec
where
    ec.id = @chunk_id
    and (
        ec.embedding_status = 'pending'
        or (ec.embedding_status = 'failed' and ec.embedding_error_count < 5)
        or (
            ec.embedding_status = 'processing'
            and ec.embedding_status_updated_at < now() - interval '2 hours'
        )
    )
    and ec.semantic_vector is null
    and ec.content is not null
    and ec.content != ''
;

-- name: UpdateChunkSemanticVector :exec
update entry_chunks
set
    semantic_vector = @semantic_vector,
    embedding_status = 'done',
    embedding_status_updated_at = now(),
    embedding_error_count = 0
where id = @chunk_id;

-- name: UpdateChunkEmbeddingStatus :exec
update entry_chunks
set
    embedding_status = @embedding_status,
    last_embedding_error = case when sqlc.narg('last_embedding_error')::text is null then last_embedding_error else @last_embedding_error end,
    last_embedding_error_at = case when sqlc.narg('last_embedding_error')::text is null then last_embedding_error_at else now() end
where id = @chunk_id;

-- name: QueryWithHybridSearch :many
with
    semantic_search as (
        select
            e.public_id,
            e.name as title,
            e.meta,
            e.entry_type as type,
            e.file_id,
            e.filesize_bytes,
            e.created_at,
            e.updated_at,
            e.archived_at,
            c.name as collection_name,
            c.public_id as collection_id,
            c.slug as collection_slug,
            w.display_name as workspace_name,
            w.public_id as workspace_id,
            w.slug as workspace_slug,
            ck.id as chunk_id,
            ck.content as chunk_content,
            ck.chunk_index,
            q.status,
            null::float8 as text_score,
            (@embedding <=> ck.semantic_vector)::float8 as semantic_score,
            rank() over (order by @embedding <=> ck.semantic_vector) as rank
        from entry_chunks ck
        join entries e on e.id = ck.entry_id
        join collections c on c.id = e.collection_id
        join workspaces w on w.id = c.workspace_id
        join entries_queue q on q.entry_id = e.id
        join workspace_members wm on wm.workspace_id = w.id
        join collection_members cm on cm.collection_id = c.id
        where
            q.status = 'completed'
            and e.archived_at is null
            and ck.semantic_vector is not null
            and (
                sqlc.narg('workspace_public_id')::uuid is null
                or w.public_id = @workspace_public_id
            )
            and (sqlc.narg('workspace_slug')::text is null or w.slug = @workspace_slug)
            and wm.user_id = @user_id
            and cm.user_id = @user_id
            and e.deleted_at is null
            and c.deleted_at is null
            and w.deleted_at is null
        order by @embedding <=> ck.semantic_vector
        limit $1
        offset $2
    ),
    text_search as (
        select
            e.public_id,
            e.name as title,
            e.meta,
            e.entry_type as type,
            e.file_id,
            e.filesize_bytes,
            e.created_at,
            e.updated_at,
            e.archived_at,
            c.name as collection_name,
            c.public_id as collection_id,
            c.slug as collection_slug,
            w.display_name as workspace_name,
            w.public_id as workspace_id,
            w.slug as workspace_slug,
            ck.id as chunk_id,
            ck.content as chunk_content,
            ck.chunk_index,
            q.status,
            ts_rank(
                ck.text_vector, websearch_to_tsquery(ts_regconfig(ck.language), @query)
            ) as text_score,
            0.0::float8 as semantic_score,
            rank() over (
                order by
                    ts_rank(
                        ck.text_vector,
                        websearch_to_tsquery(ts_regconfig(ck.language), @query)
                    )
            ) as rank
        from entry_chunks ck
        join entries e on e.id = ck.entry_id
        join collections c on c.id = e.collection_id
        join workspaces w on w.id = c.workspace_id
        join entries_queue q on q.entry_id = e.id
        join workspace_members wm on wm.workspace_id = w.id
        join collection_members cm on cm.collection_id = c.id
        where
            (ck.text_vector @@ websearch_to_tsquery(ts_regconfig(ck.language), @query))
            and q.status = 'completed'
            and e.archived_at is null
            and ck.text_vector is not null
            and (
                sqlc.narg('workspace_public_id')::uuid is null
                or w.public_id = @workspace_public_id
            )
            and (sqlc.narg('workspace_slug')::text is null or w.slug = @workspace_slug)
            and wm.user_id = @user_id
            and cm.user_id = @user_id
            and e.deleted_at is null
            and c.deleted_at is null
            and w.deleted_at is null
        order by rank
        limit $1
        offset $2
    )
select *, sum(coalesce(1.0 / (results.rank + 50), 0.0))::float8 as score
from
    (
        select *
        from semantic_search
        union all
        select *
        from text_search
    ) as results
group by
    results.public_id,
    results.title,
    results.meta,
    results.type,
    results.file_id,
    results.filesize_bytes,
    results.created_at,
    results.updated_at,
    results.archived_at,
    results.collection_name,
    results.collection_id,
    results.collection_slug,
    results.workspace_name,
    results.workspace_id,
    results.workspace_slug,
    results.chunk_id,
    results.chunk_content,
    results.chunk_index,
    results.text_score,
    results.semantic_score,
    results.rank,
    results.status
order by score desc
limit $1
offset $2
;

