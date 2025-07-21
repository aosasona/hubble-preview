ALTER TABLE entry_chunks
DROP COLUMN embeddings,
DROP COLUMN tsv;

-- Properly name the vector columns
ALTER TABLE entry_chunks
ADD COLUMN language TEXT DEFAULT 'english' NOT NULL,
ADD COLUMN text_vector tsvector DEFAULT '',
ADD COLUMN semantic_vector vector(768);

-- Create the indexes
CREATE INDEX entry_chunks_text_vector_idx ON entry_chunks USING GIN (text_vector);
CREATE INDEX entry_chunks_semantic_vector_idx ON entry_chunks USING ivfflat (semantic_vector);

-- Add comments
COMMENT ON COLUMN entry_chunks.text_vector IS 'The full-text vector of the chunk';
COMMENT ON COLUMN entry_chunks.semantic_vector IS 'The semantic vector of the chunk using an embedding model';


-- Add triggers to update the vectors
create function update_text_vector()
returns trigger
as $$
BEGIN
  NEW.text_vector := to_tsvector(NEW.language, NEW.content);
  RETURN NEW;
END;
$$
language plpgsql
;

CREATE TRIGGER entry_chunks_text_vector_update
BEFORE INSERT OR UPDATE OF content, language ON entry_chunks
FOR EACH ROW EXECUTE FUNCTION update_text_vector();

