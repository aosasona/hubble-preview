-- Drop the old trigger
drop trigger if exists entry_chunks_text_vector_update on entry_chunks;

-- Drop the old function
drop function if exists update_text_vector
;

-- Create the new function with the title prepended
create function update_text_vector()
returns trigger
as $$
BEGIN
NEW.text_vector := to_tsvector(ts_regconfig(NEW.language), NEW.content);
RETURN NEW;
END;
$$
language plpgsql
;

-- Recreate the trigger to use the new function
CREATE TRIGGER entry_chunks_text_vector_update
BEFORE INSERT OR UPDATE OF content, language ON entry_chunks
FOR EACH ROW EXECUTE FUNCTION update_text_vector();

