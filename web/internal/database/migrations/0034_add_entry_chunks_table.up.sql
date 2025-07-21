CREATE TABLE IF NOT EXISTS entry_chunks (
	id SERIAL PRIMARY KEY,
	entry_id INTEGER REFERENCES entries(id) ON DELETE CASCADE,
	chunk_index INTEGER NOT NULL CHECK (chunk_index > 0), -- the index of the chunk in the entry (for ordering)
	min_version INTEGER NOT NULL CHECK (min_version > 0), -- the minimum version of the entry that this chunk is applicable to
	content TEXT, -- the content of the chunk
	embeddings vector(768), -- the embeddings of the chunk (using nomic-embed-text's recommendation)
	tsv tsvector GENERATED ALWAYS AS (to_tsvector('english', content)) STORED,

	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ,
	deleted_at TIMESTAMPTZ DEFAULT NULL
);

