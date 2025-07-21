-- Drop the old trigger and function
DROP TRIGGER IF EXISTS entry_chunks_text_vector_update ON entry_chunks;
drop function if exists update_text_vector
;

-- Add triggers to update the vectors
create function update_text_vector()
returns trigger
as $$
BEGIN
  NEW.text_vector := to_tsvector(NEW.language::regconfig, NEW.content);
  RETURN NEW;
END;
$$
language plpgsql
;

CREATE TRIGGER entry_chunks_text_vector_update
BEFORE INSERT OR UPDATE OF content, language ON entry_chunks
FOR EACH ROW EXECUTE FUNCTION update_text_vector();

