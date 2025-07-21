-- drop existing index
DROP INDEX IF EXISTS entry_chunks_semantic_vector_idx;

CREATE INDEX ON entry_chunks USING hnsw(semantic_vector vector_cosine_ops) WITH (ef_construction=256);

