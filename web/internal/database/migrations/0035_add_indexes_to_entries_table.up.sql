-- unique constraints
CREATE UNIQUE INDEX ON entries(collection_id, slug);
CREATE UNIQUE INDEX ON entries (collection_id, origin, version);
CREATE UNIQUE INDEX ON entry_chunks(entry_id, chunk_index, min_version);

-- indexes
CREATE INDEX ON entries(collection_id);
CREATE INDEX ON entries(origin);
CREATE INDEX ON entries(parent_id);
CREATE INDEX ON entries(deleted_at);

CREATE INDEX ON entry_chunks(deleted_at);
CREATE INDEX ON entry_chunks(entry_id, chunk_index);
CREATE INDEX ON entry_chunks USING GIN(tsv);


ALTER TABLE entries
ADD CONSTRAINT ensure_valid_root_version CHECK (version = 1 OR parent_id IS NOT NULL) -- only the first version of an entry can be a root entry

